package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type repository struct {
	logger *slog.Logger
	pool   *pgxpool.Pool
}

func NewRepository(logger *slog.Logger, pool *pgxpool.Pool) *repository {
	return &repository{logger: logger, pool: pool}
}

func (r *repository) GetUnprocessedEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	const op = "repository.OutboxRepository.GetUnprocessedEvents"
	log := r.logger.With(slog.String("op", op))

	log.Info("attempt to get unprocessed events")

	query := `
		SELECT 
			id, 
      account_id, 
      transaction_id, 
      payload, 
      processed, 
      created_at 
		FROM outbox_events 
		WHERE processed = false 
		ORDER BY created_at ASC 
		LIMIT $1 
		FOR UPDATE SKIP LOCKED`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		log.Error("failed to query outbox events", "error", err)
		return nil, err
	}
	defer rows.Close()

	var events []*outboxEventRecord

	for rows.Next() {
		var e outboxEventRecord
		if err := rows.Scan(&e.ID, &e.AccountID, &e.TransactionID, &e.Payload, &e.Processed, &e.CreatedAt); err != nil {
			log.Error("failed to scan outbox event row", "error", err)
			return nil, err
		}
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return toOutboxEventsDomain(events), nil
}

func (r *repository) MarkAsProcessed(ctx context.Context, eventIDs []int64) error {
	const op = "repository.OutboxRepository.MarkAsProcessed"
	log := r.logger.With(slog.String("op", op))

	log.Info("attempt to mark events as processed", "eventIDs", eventIDs)

	query := `UPDATE outbox_events SET processed = true WHERE id = ANY($1)`
	_, err := r.pool.Exec(ctx, query, eventIDs)
	if err != nil {
		log.Error("failed to mark events as processed", "error", err)
		return err
	}

	log.Info("events marked as processed", "eventIDs", eventIDs)

	return nil
}

func toOutboxEventsDomain(events []*outboxEventRecord) []*domain.OutboxEvent {
	domainEvents := make([]*domain.OutboxEvent, 0, len(events))
	for _, event := range events {
		domainEvents = append(domainEvents, toOutboxEventDomain(event))
	}
	return domainEvents
}

func toOutboxEventDomain(e *outboxEventRecord) *domain.OutboxEvent {

	payload := domain.RestoreOutboxEventPayload(
		e.Payload.TransactionID,
		e.Payload.ExternalID,
		e.Payload.AccountID,
		e.Payload.Amount,
		e.Payload.Balance,
		e.Payload.TransactionType,
		e.Payload.Timestamp,
	)

	return domain.RestoreOutboxEvent(
		e.ID,
		e.AccountID,
		e.TransactionID,
		payload,
		e.Processed,
		e.CreatedAt,
	)
}

type outboxEventRecord struct {
	ID            int64                    `db:"id"`
	AccountID     uuid.UUID                `db:"account_id"`
	TransactionID uuid.UUID                `db:"transaction_id"`
	Payload       outboxEventRecordPayload `db:"payload"`
	Processed     bool                     `db:"processed"`
	CreatedAt     time.Time                `db:"created_at"`
}

type outboxEventRecordPayload struct {
	TransactionID   uuid.UUID              `json:"transaction_id"`
	ExternalID      uuid.UUID              `json:"external_id"`
	AccountID       uuid.UUID              `json:"account_id"`
	Amount          decimal.Decimal        `json:"amount"`
	Balance         decimal.Decimal        `json:"balance"`
	TransactionType domain.TransactionType `json:"transaction_type"`
	Timestamp       time.Time              `json:"timestamp"`
}

func (p *outboxEventRecordPayload) Scan(src any) error {

	if src == nil {
		return nil
	}

	data, ok := src.([]byte)
	if !ok {
		return errors.New("expected []byte, got " + fmt.Sprintf("%T", src))
	}

	return json.Unmarshal(data, p)
}
