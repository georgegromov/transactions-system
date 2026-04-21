package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/georgegromov/transactions-system/services/balance-viewer/internal/domain"
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

func (r *repository) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (*domain.Balance, error) {
	const op = "repository.BalanceRepository.GetAccountBalance"
	log := r.logger.With(slog.String("op", op), slog.String("account_id", accountID.String()))

	query := `SELECT account_id, current_balance, updated_at FROM balances WHERE account_id = $1`

	var balanceRecord balanceRecord
	if err := r.pool.QueryRow(ctx, query, accountID).
		Scan(
			&balanceRecord.AccountID,
			&balanceRecord.CurrentBalance,
			&balanceRecord.UpdatedAt,
		); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountBalanceNotFound
		}
		log.Error("failed to get account balance", "error", err)
		return nil, err
	}

	return toDomainBalance(balanceRecord), nil
}

type balanceRecord struct {
	AccountID      uuid.UUID       `db:"account_id"`
	CurrentBalance decimal.Decimal `db:"current_balance"`
	UpdatedAt      time.Time       `db:"updated_at"`
}

func toDomainBalance(record balanceRecord) *domain.Balance {
	return domain.RestoreBalance(
		record.AccountID,
		record.CurrentBalance,
		record.UpdatedAt,
	)
}
