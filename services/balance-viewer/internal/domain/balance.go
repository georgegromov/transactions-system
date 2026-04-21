package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrAccountBalanceNotFound = errors.New("account balance not found")
)

type Balance struct {
	accountID      uuid.UUID
	currentBalance decimal.Decimal
	updatedAt      time.Time
}

func (b *Balance) AccountID() uuid.UUID            { return b.accountID }
func (b *Balance) CurrentBalance() decimal.Decimal { return b.currentBalance }
func (b *Balance) UpdatedAt() time.Time            { return b.updatedAt }

func RestoreBalance(
	accountID uuid.UUID,
	currentBalance decimal.Decimal,
	updatedAt time.Time,
) *Balance {
	return &Balance{
		accountID:      accountID,
		currentBalance: currentBalance,
		updatedAt:      updatedAt,
	}
}
