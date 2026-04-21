package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	ErrInsufficientBalance      = errors.New("insufficient balance")
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

func (t TransactionType) String() string {
	return string(t)
}

type Transaction struct {
	id              uuid.UUID
	externalID      uuid.UUID
	accountID       uuid.UUID
	amount          decimal.Decimal
	balance         decimal.Decimal
	transactionType TransactionType
	createdAt       time.Time
}

// Getters
func (t *Transaction) ID() uuid.UUID                    { return t.id }
func (t *Transaction) ExternalID() uuid.UUID            { return t.externalID }
func (t *Transaction) AccountID() uuid.UUID             { return t.accountID }
func (t *Transaction) Amount() decimal.Decimal          { return t.amount }
func (t *Transaction) Balance() decimal.Decimal         { return t.balance }
func (t *Transaction) TransactionType() TransactionType { return t.transactionType }
func (t *Transaction) CreatedAt() time.Time             { return t.createdAt }

func RestoreTransaction(
	id, externalID, accountID uuid.UUID,
	amount, balance decimal.Decimal,
	transactionType TransactionType,
	createdAt time.Time,
) *Transaction {
	return &Transaction{
		id:              id,
		externalID:      externalID,
		accountID:       accountID,
		amount:          amount,
		balance:         balance,
		transactionType: transactionType,
		createdAt:       createdAt,
	}
}
