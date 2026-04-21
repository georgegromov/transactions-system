package domain

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	id            int64
	accountID     uuid.UUID
	transactionID uuid.UUID
	payload       OutboxEventPayload
	processed     bool
	createdAt     time.Time
}

func (e *OutboxEvent) ID() int64                   { return e.id }
func (e *OutboxEvent) AccountID() uuid.UUID        { return e.accountID }
func (e *OutboxEvent) TransactionID() uuid.UUID    { return e.transactionID }
func (e *OutboxEvent) Payload() OutboxEventPayload { return e.payload }
func (e *OutboxEvent) Processed() bool             { return e.processed }
func (e *OutboxEvent) CreatedAt() time.Time        { return e.createdAt }

func RestoreOutboxEvent(
	id int64,
	accountID, transactionID uuid.UUID,
	payload OutboxEventPayload,
	processed bool,
	createdAt time.Time,
) *OutboxEvent {
	return &OutboxEvent{
		id:            id,
		accountID:     accountID,
		transactionID: transactionID,
		payload:       payload,
		processed:     processed,
		createdAt:     createdAt,
	}
}
