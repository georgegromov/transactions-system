package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

func (t TransactionType) String() string {
	return string(t)
}

type OutboxEventPayload struct {
	transactionID   uuid.UUID
	externalID      uuid.UUID
	accountID       uuid.UUID
	amount          decimal.Decimal
	balance         decimal.Decimal
	transactionType TransactionType
	timestamp       time.Time
}

func (p *OutboxEventPayload) TransactionID() uuid.UUID         { return p.transactionID }
func (p *OutboxEventPayload) ExternalID() uuid.UUID            { return p.externalID }
func (p *OutboxEventPayload) AccountID() uuid.UUID             { return p.accountID }
func (p *OutboxEventPayload) Amount() decimal.Decimal          { return p.amount }
func (p *OutboxEventPayload) Balance() decimal.Decimal         { return p.balance }
func (p *OutboxEventPayload) TransactionType() TransactionType { return p.transactionType }
func (p *OutboxEventPayload) Timestamp() time.Time             { return p.timestamp }

func RestoreOutboxEventPayload(
	transactionID, externalID, accountID uuid.UUID,
	amount, balance decimal.Decimal,
	transactionType TransactionType,
	timestamp time.Time,
) OutboxEventPayload {
	return OutboxEventPayload{
		transactionID:   transactionID,
		externalID:      externalID,
		accountID:       accountID,
		amount:          amount,
		balance:         balance,
		transactionType: transactionType,
		timestamp:       timestamp,
	}
}
