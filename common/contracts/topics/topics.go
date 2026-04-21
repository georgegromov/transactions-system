package topics

type Topic string

const (
	TopicTransactions Topic = "transactions"
)

func (t Topic) String() string {
	return string(t)
}
