package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/georgegromov/transactions-system/common/consul"
	"github.com/georgegromov/transactions-system/common/db"
	"github.com/georgegromov/transactions-system/common/httpserver"
	"github.com/georgegromov/transactions-system/common/logger"
	"github.com/georgegromov/transactions-system/common/validation"
	"github.com/georgegromov/transactions-system/services/transaction-processor/internal/config"
	"github.com/georgegromov/transactions-system/services/transaction-processor/internal/repository"
	"github.com/georgegromov/transactions-system/services/transaction-processor/internal/service"
	"github.com/georgegromov/transactions-system/services/transaction-processor/internal/transport/http"
)

const (
	consulHTTPAddr   string = "http://consul:8500"
	consulHealthPath string = "/health"

	serviceID   string = "transaction-processor-1"
	serviceName string = "transaction-processor"
	serviceHost string = "transaction-processor"

	shutdownTimeout time.Duration = 10 * time.Second
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

	validator := validation.NewService()

	transactionRepository := repository.NewTransactionRepository(log, pool.Pool())
	transactionService := service.NewTransactionService(log, transactionRepository)
	transactionsHandler := http.NewTransactionsHandler(log, validator, transactionService)

	router := http.NewRouter(http.RouterDeps{
		TransactionsHandler: transactionsHandler,
	})

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

	log.Info("shutting down service...", slog.String("reason", ctx.Err().Error()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Error("failed to stop http server gracefully", "error", err)
		panic(err)
	}

	log.Info("service stopped gracefully")
}
