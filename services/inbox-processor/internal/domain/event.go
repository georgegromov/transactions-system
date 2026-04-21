package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInboxEventAlreadyExists = errors.New("inbox event already exists")
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

func (t TransactionType) String() string {
	return string(t)
}

type InboxEvent struct {
	transactionID   uuid.UUID
	externalID      uuid.UUID
	accountID       uuid.UUID
	amount          decimal.Decimal
	balance         decimal.Decimal
	transactionType TransactionType
	timestamp       time.Time
}

func (e *InboxEvent) TransactionID() uuid.UUID         { return e.transactionID }
func (e *InboxEvent) ExternalID() uuid.UUID            { return e.externalID }
func (e *InboxEvent) AccountID() uuid.UUID             { return e.accountID }
func (e *InboxEvent) Amount() decimal.Decimal          { return e.amount }
func (e *InboxEvent) Balance() decimal.Decimal         { return e.balance }
func (e *InboxEvent) TransactionType() TransactionType { return e.transactionType }
func (e *InboxEvent) Timestamp() time.Time             { return e.timestamp }

func RestoreInboxEvent(
	transactionID, externalID, accountID uuid.UUID,
	amount, balance decimal.Decimal,
	transactionType TransactionType,
	timestamp time.Time,
) *InboxEvent {
	return &InboxEvent{
		transactionID:   transactionID,
		externalID:      externalID,
		accountID:       accountID,
		amount:          amount,
		balance:         balance,
		transactionType: transactionType,
		timestamp:       timestamp,
	}
}
