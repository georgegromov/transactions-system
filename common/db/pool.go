package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	minConns        = 5
	maxConns        = 20
	maxConnLifetime = time.Hour
	maxConnIdleTime = time.Minute * 30
)

type pgPool struct {
	pool *pgxpool.Pool
}

func NewPool(ctx context.Context, dsn string) (*pgPool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	config.MinConns = minConns
	config.MaxConns = maxConns
	config.MaxConnLifetime = maxConnLifetime
	config.MaxConnIdleTime = maxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &pgPool{pool: pool}, nil
}

func (p *pgPool) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *pgPool) Close() {
	p.pool.Close()
}
