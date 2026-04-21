package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/georgegromov/transactions-system/services/transaction-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type transactionRepository struct {
	logger *slog.Logger
	pool   *pgxpool.Pool
}

func NewTransactionRepository(
	logger *slog.Logger,
	pool *pgxpool.Pool,
) *transactionRepository {
	return &transactionRepository{
		logger: logger,
		pool:   pool,
	}
}

func (r *transactionRepository) CreateTransaction(
	ctx context.Context,
	externalID, accountID uuid.UUID,
	amount decimal.Decimal,
	transactionType domain.TransactionType,
) error {
	const op = "repository.TransactionRepository.CreateTransaction"
	log := r.logger.With(slog.String("op", op), slog.String("account_id", accountID.String()))

	log.Info("attempt to create transaction")

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	safeAmount := amount
	if transactionType == domain.TransactionTypeExpense && safeAmount.Sign() > 0 {
		safeAmount = safeAmount.Neg()
	}

	insertBalanceQuery := `
  INSERT INTO balances (account_id, current_balance, updated_at)
  VALUES ($1, $2, NOW())
  ON CONFLICT (account_id) DO UPDATE SET 
    current_balance = balances.current_balance + EXCLUDED.current_balance, 
    updated_at = NOW()
  RETURNING current_balance
  `

	var currentBalance decimal.Decimal
	if err := tx.QueryRow(
		ctx,
		insertBalanceQuery,
		accountID,
		safeAmount,
	).Scan(&currentBalance); err != nil {
		log.Error("failed to insert balance", "error", err)
		return err
	}

	if currentBalance.Sign() < 0 {
		return domain.ErrInsufficientBalance
	}

	insertTransactionQuery := `
    INSERT INTO transactions (external_id, account_id, amount, balance, transaction_type)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, created_at
  `

	var transactionID uuid.UUID
	var transactionCreatedAt time.Time

	if err := tx.QueryRow(
		ctx,
		insertTransactionQuery,
		externalID,
		accountID,
		amount,
		currentBalance,
		transactionType,
	).Scan(&transactionID, &transactionCreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Warn("transaction with this external_id already exists")
			return domain.ErrTransactionAlreadyExists
		}

		log.Error("failed to insert transaction", "error", err)
		return err
	}

	insertOutboxEventQuery := `
    INSERT INTO outbox_events (account_id, transaction_id, payload)
    VALUES ($1, $2, $3)
  `

	payload := transactionPayload{
		TransactionID:   transactionID,
		ExternalID:      externalID,
		AccountID:       accountID,
		Amount:          amount,
		Balance:         currentBalance,
		TransactionType: transactionType,
		Timestamp:       transactionCreatedAt,
	}

	if _, err := tx.Exec(
		ctx, insertOutboxEventQuery,
		accountID,
		transactionID,
		payload,
	); err != nil {
		log.Error("failed to insert outbox event", "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("an error occurred while committing the transaction", "error", err)
		return err
	}

	return nil
}

type transactionPayload struct {
	TransactionID   uuid.UUID              `json:"transaction_id"`
	ExternalID      uuid.UUID              `json:"external_id"`
	AccountID       uuid.UUID              `json:"account_id"`
	Amount          decimal.Decimal        `json:"amount"`
	Balance         decimal.Decimal        `json:"balance"`
	TransactionType domain.TransactionType `json:"transaction_type"`
	Timestamp       time.Time              `json:"timestamp"`
}

func (p transactionPayload) Value() (driver.Value, error) {
	return json.Marshal(p)
}
