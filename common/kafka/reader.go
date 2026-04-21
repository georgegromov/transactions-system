package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	startOffset    int64         = kafka.FirstOffset
	commitInterval time.Duration = 0
	minBytes       int           = 1
	maxBytes       int           = 10e6 // 10 MB
	maxWait        time.Duration = 500 * time.Millisecond
)

type reader struct {
	reader *kafka.Reader
}

func NewReader(brokers []string, groupID, topic string) *reader {
	return &reader{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,

			// Откуда читать при первом запуске
			StartOffset: startOffset,

			// Отключаем автокоммит - коммитим вручную после обработки
			CommitInterval: commitInterval,

			// Минимум и максимум данных за один fetch
			MinBytes: minBytes,
			MaxBytes: maxBytes,

			// Таймаут ожидания новых сообщений
			MaxWait: maxWait,

			// Logger
			Logger:      kafka.LoggerFunc(func(msg string, args ...interface{}) { slog.Debug(fmt.Sprintf(msg, args...)) }),
			ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) { slog.Error(fmt.Sprintf(msg, args...)) }),
		}),
	}
}

func (r *reader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return r.reader.FetchMessage(ctx)
}

func (r *reader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return r.reader.CommitMessages(ctx, msgs...)
}

func (r *reader) Close() error {
	return r.reader.Close()
}
