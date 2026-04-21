package logger

import (
	"fmt"
	"log/slog"
	"os"
)

type env string

const (
	envLocal env = "local"
	envDev   env = "dev"
	envProd  env = "prod"
)

func (e env) IsValid() bool {
	return e == envLocal || e == envDev || e == envProd
}

func parseEnv(envStr string) env {
	e := env(envStr)
	if !e.IsValid() {
		panic(fmt.Errorf("invalid environment: %s", envStr))
	}

	return e
}

func NewLogger(env string) *slog.Logger {
	e := parseEnv(env)

	var level slog.Level

	switch e {
	case envProd:
		level = slog.LevelInfo
	default:
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
