package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type writer struct {
	writer *kafka.Writer
}

func NewWriter(brokers []string) *writer {
	return &writer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.Hash{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (w *writer) WriteMessages(ctx context.Context, messages ...kafka.Message) error {
	return w.writer.WriteMessages(ctx, messages...)
}

func (w *writer) WriteMessage(ctx context.Context, topic, key string, payload []byte) error {
	return w.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	},
	)
}

func (w *writer) Close() error {
	return w.writer.Close()
}
