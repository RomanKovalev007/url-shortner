package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	FlagPostgres = "postgres"
	FlagInMemory = "in-memory"
)

type Config struct {
	BaseURL  string `env:"BASE_URL"   env-default:"localhost"`
	DbFlag   string `env:"DB_FLAG"    env-required:"true"`
	LogLevel string `env:"LOG_LEVEL"  env-default:"info"`
	HTTP     HTTP
	Postgres PostgresConfig
	InMemory InMemoryConfig
}

type InMemoryConfig struct {
	TTL             time.Duration `env:"INMEMORY_TTL"              env-default:"0"`
	CleanupInterval time.Duration `env:"INMEMORY_CLEANUP_INTERVAL" env-default:"1h"`
}

type HTTP struct {
	Addr         string        `env:"HTTP_ADDR"          env-default:":8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"  env-default:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT"  env-default:"60s"`
}

type PostgresConfig struct {
	Host            string `env:"POSTGRES_HOST"            env-default:"db"`
	Port            string `env:"POSTGRES_PORT"            env-default:"5432"`
	User            string `env:"POSTGRES_USER"            env-default:"postgres"`
	Password        string `env:"POSTGRES_PASSWORD"        env-default:"postgres"`
	DB              string `env:"POSTGRES_DB"              env-default:"urls"`
	SSLMode         string `env:"POSTGRES_SSLMODE"         env-default:"disable"`
	MaxConns        int32  `env:"POSTGRES_MAXCONNS"        env-default:"20"`
	MinConns        int32  `env:"POSTGRES_MINCONNS"        env-default:"2"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAXCONNLIFETIME" env-default:"30m"`
	MaxConnIdleTime time.Duration `env:"POSTGRES_MAXCONNIDLE"     env-default:"5m"`
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DB, c.SSLMode)
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return &cfg, nil
}
