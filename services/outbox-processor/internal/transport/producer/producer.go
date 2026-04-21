package producer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

type kafkaWriter interface {
	WriteMessages(ctx context.Context, messages ...kafka.Message) error
}

type outboxEventProducer struct {
	logger *slog.Logger
	writer kafkaWriter
	topic  string
}

func NewEventProducer(logger *slog.Logger, writer kafkaWriter, topic string) *outboxEventProducer {
	return &outboxEventProducer{
		logger: logger,
		writer: writer,
		topic:  topic,
	}
}

func (p *outboxEventProducer) ProduceTransactionCreatedEvent(ctx context.Context, events []*domain.OutboxEvent) error {
	const op = "outboxEventProducer.ProduceTransactionCreatedEvent"
	log := p.logger.With(slog.String("op", op))

	msgs := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		payload := event.Payload()

		msg := transactionCreatedEventMessage{
			TransactionID:   payload.TransactionID(),
			ExternalID:      payload.ExternalID(),
			AccountID:       payload.AccountID(),
			Amount:          payload.Amount(),
			Balance:         payload.Balance(),
			TransactionType: payload.TransactionType(),
			Timestamp:       payload.Timestamp(),
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Error("failed to marshal transaction created event message", "error", err)
			return err
		}

		key := payload.AccountID().String()

		msgs = append(msgs, kafka.Message{
			Topic: p.topic,
			Key:   []byte(key),
			Value: msgBytes,
		})

	}

	if err := p.writer.WriteMessages(ctx, msgs...); err != nil {
		log.Error("failed to write transaction created event messages", "error", err)
		return err
	}

	return nil
}

type transactionCreatedEventMessage struct {
	TransactionID   uuid.UUID              `json:"transaction_id"`
	ExternalID      uuid.UUID              `json:"external_id"`
	AccountID       uuid.UUID              `json:"account_id"`
	Amount          decimal.Decimal        `json:"amount"`
	Balance         decimal.Decimal        `json:"balance"`
	TransactionType domain.TransactionType `json:"transaction_type"`
	Timestamp       time.Time              `json:"timestamp"`
}
