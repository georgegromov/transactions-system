package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/georgegromov/transactions-system/services/inbox-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

// Интерфейс для взаимодействия с нашим common/kafka/reader.go
type kafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

type eventConsumer struct {
	logger *slog.Logger
	reader kafkaReader
}

func NewEventConsumer(logger *slog.Logger, reader kafkaReader) *eventConsumer {
	return &eventConsumer{
		logger: logger,
		reader: reader,
	}
}

// ReadTransactionEvent читает сообщение из Kafka и возвращает доменный объект и функцию для коммита
func (c *eventConsumer) ReadTransactionEvent(ctx context.Context) (*domain.InboxEvent, func() error, error) {

	const op = "eventConsumer.ReadTransactionEvent"
	log := c.logger.With(slog.String("op", op))

	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		log.Error("failed to fetch kafka message", "error", err)
		return nil, nil, err
	}

	log.Debug("received message from kafka", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)

	// 2. Парсим JSON в DTO
	var dto transactionCreatedEventMessage
	if err := json.Unmarshal(msg.Value, &dto); err != nil {
		log.Error("failed to unmarshal kafka message", "error", err, "payload", string(msg.Value))
		// Если пришел битый JSON, мы не можем его обработать.
		// В реальном проде тут часто делают Commit, чтобы битое сообщение не заблокировало очередь навсегда.
		// Но пока просто возвращаем ошибку.
		return nil, nil, err
	}

	// 3. Собираем доменную модель
	event := domain.RestoreInboxEvent(
		dto.TransactionID,
		dto.ExternalID,
		dto.AccountID,
		dto.Amount,
		dto.Balance,
		dto.TransactionType,
		dto.Timestamp,
	)

	// 4. Создаем замыкание (closure) для коммита.
	// Когда воркер вызовет эту функцию, она подтвердит именно ЭТО сообщение.
	commitFunc := func() error {
		if err := c.reader.CommitMessages(context.Background(), msg); err != nil {
			log.Error("failed to commit kafka message", "error", err, "offset", msg.Offset)
			return err
		}
		log.Debug("kafka message committed", "offset", msg.Offset)
		return nil
	}

	return event, commitFunc, nil
}

// DTO, в который мы будем парсить JSON из Кафки
// Он должен полностью совпадать с тем, что отправлял outbox-processor!
type transactionCreatedEventMessage struct {
	TransactionID   uuid.UUID              `json:"transaction_id"`
	ExternalID      uuid.UUID              `json:"external_id"`
	AccountID       uuid.UUID              `json:"account_id"`
	Amount          decimal.Decimal        `json:"amount"`
	Balance         decimal.Decimal        `json:"balance"`
	TransactionType domain.TransactionType `json:"transaction_type"`
	Timestamp       time.Time              `json:"timestamp"`
}
