package main

import (
	"context"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/georgegromov/transactions-system/common/consul"
	"github.com/georgegromov/transactions-system/common/contracts/topics"
	"github.com/georgegromov/transactions-system/common/db"
	"github.com/georgegromov/transactions-system/common/httpserver"
	"github.com/georgegromov/transactions-system/common/kafka"
	"github.com/georgegromov/transactions-system/common/logger"
	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/config"
	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/repository"
	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/transport/http"
	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/transport/producer"
	"github.com/georgegromov/transactions-system/services/outbox-processor/internal/worker"
)

const (
	consulHTTPAddr   string = "http://consul:8500"
	consulHealthPath string = "/health"

	serviceID   string = "outbox-processor-1"
	serviceName string = "outbox-processor"
	serviceHost string = "outbox-processor"

	shutdownTimeout time.Duration = 10 * time.Second

	pollInterval time.Duration = time.Second * 1
	batchSize    int           = 100
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()

	logger := logger.NewLogger(cfg.Env)
	log := logger.With(slog.String("service", serviceName), slog.String("env", cfg.Env))
	log.Info("starting service...")

	pool, err := db.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.Error("failed to create database connection pool", "error", err)
		panic(err)
	}
	defer pool.Close()

	writer := kafka.NewWriter(cfg.Kafka.Brokers)
	defer func() {
		log.Info("closing kafka writer...")
		if err := writer.Close(); err != nil {
			log.Error("failed to close kafka writer", "error", err)
		}
	}()

	outboxEventProducer := producer.NewEventProducer(log, writer, topics.TopicTransactions.String())
	outboxRepository := repository.NewRepository(log, pool.Pool())
	outboxProcessor := worker.NewOutboxProcessor(log, outboxRepository, outboxEventProducer, pollInterval, batchSize)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		outboxProcessor.Start(ctx)
	}()

	router := http.NewRouter()

	server := httpserver.NewHttpServer(
		cfg.Http.Port,
		cfg.Http.MaxHeaderBytes,
		cfg.Http.ReadTimeout,
		cfg.Http.WriteTimeout,
		cfg.Http.IdleTimeout,
		router,
	)

	deregister, err := consul.RegisterMaybe(consul.Config{
		ConsulHTTPAddr: consulHTTPAddr,
		ServiceID:      serviceID,
		ServiceName:    serviceName,
		ServiceHost:    serviceHost,
		Port:           cfg.Http.Port,
		HealthPath:     consulHealthPath,
	})
	if err != nil {
		log.Error("failed to register service in consul", "error", err)
		panic(err)
	}
	defer deregister()

	go func() {
		log.Info("starting http server", slog.Int("port", cfg.Http.Port))
		if err := server.Start(); err != nil {
			log.Error("failed to start http server", slog.Int("port", cfg.Http.Port), "error", err)
			panic(err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Error("failed to stop http server gracefully", "error", err)
	}

	log.Info("waiting for outbox processor to finish...")
	wg.Wait()

	log.Info("service stopped gracefully")
}
