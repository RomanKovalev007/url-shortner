package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	BaseURl  string `env:"BASE_URL"   env-default:"localhost"`
	DbFlag   string `env:"DB_FLAG"    env-default:"postgres"`
	LogLevel string `env:"LOG_LEVEL"  env-default:"info"`
	HTTP HTTP
}

type HTTP struct {
	Addr         string        `env:"HTTP_ADDR"          env-default:":8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"  env-default:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT"  env-default:"60s"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return &cfg, nil
}
