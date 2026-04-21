package service

import (
	"context"
	"log/slog"

	"github.com/georgegromov/transactions-system/services/balance-viewer/internal/domain"
	"github.com/google/uuid"
)

type balanceRepository interface {
	GetAccountBalance(ctx context.Context, accountID uuid.UUID) (*domain.Balance, error)
}

type service struct {
	logger            *slog.Logger
	balanceRepository balanceRepository
}

func NewService(logger *slog.Logger, balanceRepository balanceRepository) *service {
	return &service{logger: logger, balanceRepository: balanceRepository}
}

func (s *service) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (*domain.Balance, error) {
	return s.balanceRepository.GetAccountBalance(ctx, accountID)
}
