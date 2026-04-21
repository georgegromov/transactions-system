package repository

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/georgegromov/transactions-system/services/inbox-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *repository) ProcessTransaction(ctx context.Context, event *domain.InboxEvent) error {
	const op = "repository.InboxRepository.ProcessTransaction"
	log := r.logger.With(slog.String("op", op))

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	insertInboxQuery := `
    INSERT INTO inbox_events (account_id, transaction_id, payload)
    VALUES ($1, $2, $3)
    ON CONFLICT (transaction_id) DO NOTHING
  `

	payloadData := inboxEventPayload{
		TransactionID:   event.TransactionID(),
		ExternalID:      event.ExternalID(),
		AccountID:       event.AccountID(),
		Amount:          event.Amount(),
		Balance:         event.Balance(),
		TransactionType: event.TransactionType(),
		Timestamp:       event.Timestamp(),
	}

	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		log.Error("failed to marshal payload", "error", err)
		return err
	}

	tag, err := tx.Exec(ctx, insertInboxQuery, event.AccountID(), event.TransactionID(), payloadBytes)
	if err != nil {
		log.Error("failed to insert inbox event", "error", err)
		return err
	}

	if tag.RowsAffected() == 0 {
		log.Info("transaction already processed, skipping (idempotency triggered)")
		return nil
	}

	safeAmount := event.Amount()
	if event.TransactionType() == domain.TransactionTypeExpense && safeAmount.Sign() > 0 {
		safeAmount = safeAmount.Neg()
	}

	updateBalanceQuery := `
    INSERT INTO balances (account_id, current_balance, updated_at)
    VALUES ($1, $2, NOW())
    ON CONFLICT (account_id) DO UPDATE SET 
      current_balance = balances.current_balance + EXCLUDED.current_balance, 
      updated_at = NOW()
  `
	if _, err := tx.Exec(ctx, updateBalanceQuery, event.AccountID(), safeAmount); err != nil {
		log.Error("failed to update balance", "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", "error", err)
		return err
	}
	log.Info("successfully processed transaction and updated balance")
	return nil
}

type inboxEventPayload struct {
	TransactionID   uuid.UUID              `json:"transaction_id"`
	ExternalID      uuid.UUID              `json:"external_id"`
	AccountID       uuid.UUID              `json:"account_id"`
	Amount          decimal.Decimal        `json:"amount"`
	Balance         decimal.Decimal        `json:"balance"`
	TransactionType domain.TransactionType `json:"transaction_type"`
	Timestamp       time.Time              `json:"timestamp"`
}
