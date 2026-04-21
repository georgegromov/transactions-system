package worker

import (
	"context"
	"log/slog"

	"github.com/georgegromov/transactions-system/services/inbox-processor/internal/domain"
)

type repository interface {
	ProcessTransaction(ctx context.Context, event *domain.InboxEvent) error
}

type eventConsumer interface {
	ReadTransactionEvent(ctx context.Context) (*domain.InboxEvent, func() error, error)
}

type inboxProcessor struct {
	logger     *slog.Logger
	repository repository
	consumer   eventConsumer
}

func NewInboxProcessor(logger *slog.Logger, repository repository, consumer eventConsumer) *inboxProcessor {
	return &inboxProcessor{logger: logger, repository: repository, consumer: consumer}
}

func (p *inboxProcessor) Start(ctx context.Context) {
	const op = "inboxProcessor.Start"
	log := p.logger.With(slog.String("op", op))
	log.Info("starting inbox worker...")
	for {
		// 1. Проверяем, не просили ли нас остановиться (Graceful Shutdown)
		if ctx.Err() != nil {
			log.Info("context canceled, stopping inbox worker")
			return
		}
		// 2. Читаем сообщение из Kafka (этот вызов заблокируется, пока не придет сообщение)
		event, commitFunc, err := p.consumer.ReadTransactionEvent(ctx)
		if err != nil {
			// Если контекст отменили во время чтения, это ок
			if ctx.Err() != nil {
				return
			}

			// Если пришел битый JSON или отвалилась Кафка - логируем и идем дальше
			// (В реальном проекте битые JSON иногда коммитят сразу, чтобы не застрять на них навсегда)
			log.Error("failed to read transaction event from kafka", "error", err)
			continue
		}
		// 3. Передаем событие в Репозиторий
		// Репозиторий сам разберется: это новая транзакция или дубликат (DO NOTHING)
		err = p.repository.ProcessTransaction(ctx, event)
		if err != nil {
			// КРИТИЧЕСКИЙ МОМЕНТ:
			// Если база данных упала, мы НЕ ВЫЗЫВАЕМ commitFunc()!
			// Воркер пойдет на следующий круг, и Кафка отдаст ему это же сообщение еще раз.
			// Так мы никогда не потеряем деньги клиента.
			log.Error("failed to process transaction in db", "error", err, "transaction_id", event.TransactionID())
			continue
		}
		// 4. Если всё прошло успешно (записали в БД или проигнорировали дубль),
		// мы обязаны сказать Кафке: "Всё ок, сдвигай оффсет!"
		if err := commitFunc(); err != nil {
			log.Error("failed to commit kafka message", "error", err)
			// Даже если коммит не прошел, мы продолжаем работу.
			// Кафка потом пришлет нам это сообщение снова, но наш Репозиторий
			// поймает его через ON CONFLICT DO NOTHING (идемпотентность).
		}
	}
}
