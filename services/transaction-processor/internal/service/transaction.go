package service

import (
	"context"
	"log/slog"

	"github.com/georgegromov/transactions-system/services/transaction-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type transactionRepository interface {
	CreateTransaction(ctx context.Context, externalID, accountID uuid.UUID, amount decimal.Decimal, transactionType domain.TransactionType) error
}

type transactionService struct {
	logger                *slog.Logger
	transactionRepository transactionRepository
}

func NewTransactionService(
	logger *slog.Logger,
	transactionRepository transactionRepository,
) *transactionService {
	return &transactionService{
		logger:                logger,
		transactionRepository: transactionRepository,
	}
}

func (s *transactionService) CreateTransaction(
	ctx context.Context,
	externalID, accountID uuid.UUID,
	amount decimal.Decimal,
	transactionType domain.TransactionType,
) error {
	return s.transactionRepository.CreateTransaction(ctx, externalID, accountID, amount, transactionType)
}
