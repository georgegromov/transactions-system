package config

import (
	"fmt"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `env:"ENV" env-default:"local"`
	Http     Http     `env-prefix:"BALANCE_VIEWER_HTTP_"`
	Database Database `env-prefix:"BALANCES_DB_"`
}

type Http struct {
	Port           int           `env:"PORT" env-required:"true"`
	ReadTimeout    time.Duration `env:"READ_TIMEOUT" env-required:"true"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" env-required:"true"`
	IdleTimeout    time.Duration `env:"IDLE_TIMEOUT" env-required:"true"`
	MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" env-required:"true"`
}

type Database struct {
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Name     string `env:"POSTGRES_DB" env-required:"true"`
	Port     int    `env:"POSTGRES_PORT" env-required:"true"`
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
}

func (db Database) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", db.User, db.Password, db.Host, db.Port, db.Name)
}

func MustLoad() *Config {
	var cfg Config
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			log.Fatalf("cannot read config: %v", err)
		}
	}
	return &cfg
}
