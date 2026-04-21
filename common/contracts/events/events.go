package events

type EventType string

const (
	EventTransactionCreated EventType = "transaction.created"
)

func (e EventType) String() string {
	return string(e)
}
