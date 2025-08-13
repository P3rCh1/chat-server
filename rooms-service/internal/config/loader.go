package config

import (
	"errors"
	"os"
	"time"

	"github.com/P3rCh1/chat-server/rooms-service/shared/config"
)

type Config struct {
	Port            string        `yaml:"port"`
	LogLevel        string        `yaml:"log_level"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	Postgres        *Postgres     `yaml:"postgres"`
	Redis           *Redis        `yaml:"redis"`
}

type Postgres struct {
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	DB       string `yaml:"db"`
	User     string `yaml:"user"`
	Password string
}

type Redis struct {
	Addr     string        `yaml:"addr"`
	DB       int           `yaml:"db"`
	TTL      time.Duration `yaml:"ttl"`
	Password string
}

func (cfg *Config) Validate() error {
	if cfg.Postgres.Password == "" {
		return errors.New("postgres password is required")
	}
	if cfg.Redis.Password == "" {
		return errors.New("redis password is required")
	}
	return nil
}

func Default() *Config {
	return &Config{
		LogLevel:        "info",
		Port:            ":50053",
		ShutdownTimeout: 10 * time.Second,
		Postgres: &Postgres{
			Port: "5432",
			Host: "postgres",
			DB:   "chatdb",
			User: "chat",
		},
		Redis: &Redis{
			Addr: "redis:6379",
			TTL:  24 * time.Hour,
		},
	}
}

func MustLoad() *Config {
	cfg := Default()
	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	config.MustLoad(cfg)
	return cfg
}
