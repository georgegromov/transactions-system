package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/domain"
)

type repository interface {
	GetUnprocessedEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)
	MarkAsProcessed(ctx context.Context, eventIDs []int64) error
}

type eventProducer interface {
	ProduceTransactionCreatedEvent(ctx context.Context, events []*domain.OutboxEvent) error
}

type outboxProcessor struct {
	// dependencies
	logger        *slog.Logger
	repository    repository
	eventProducer eventProducer

	// config
	pollInterval time.Duration
	batchSize    int
}

func NewOutboxProcessor(
	logger *slog.Logger,
	repository repository,
	eventProducer eventProducer,
	pollInterval time.Duration,
	batchSize int,
) *outboxProcessor {
	return &outboxProcessor{
		logger:        logger,
		repository:    repository,
		eventProducer: eventProducer,
		pollInterval:  pollInterval,
		batchSize:     batchSize,
	}
}

func (p *outboxProcessor) Start(ctx context.Context) {
	const op = "outboxProcessor.Start"
	log := p.logger.With(slog.String("op", op))

	log.Info("starting outbox processor", slog.String("poll_interval", p.pollInterval.String()), slog.Int("batch_size", p.batchSize))

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("context done, stopping outbox processor")
			return
		case <-ticker.C:
			for {

				if ctx.Err() != nil {
					log.Info("context done, stopping outbox processor")
					return
				}

				eventCount, err := p.processEvents(ctx)
				if err != nil {
					break
				}

				if eventCount < p.batchSize {
					break
				}
			}
		}
	}
}

func (p *outboxProcessor) processEvents(ctx context.Context) (int, error) {
	const op = "outboxProcessor.processEvents"
	log := p.logger.With(slog.String("op", op))

	log.Debug("attempt to process events")

	events, err := p.repository.GetUnprocessedEvents(ctx, p.batchSize)
	if err != nil {
		log.Error("failed to get unprocessed events", "error", err)
		return 0, err
	}

	eventCount := len(events)

	if eventCount == 0 {
		log.Debug("no unprocessed events found")
		return 0, nil
	}

	if err := p.eventProducer.ProduceTransactionCreatedEvent(ctx, events); err != nil {
		log.Error("failed to produce events batch", "error", err)
		return 0, err
	}

	eventIDs := make([]int64, 0, eventCount)
	for _, event := range events {
		eventIDs = append(eventIDs, event.ID())
	}

	if err := p.repository.MarkAsProcessed(ctx, eventIDs); err != nil {
		log.Error("failed to mark events as processed", "error", err, "eventIDs", eventIDs)
		return 0, err
	}

	log.Info("successfully processed outbox events", "count", eventCount)

	return eventCount, nil
}
