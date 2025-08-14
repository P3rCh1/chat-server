package config

import (
	"errors"
	"os"
	"time"

	"github.com/P3rCh1/chat-server/message-service/pkg/config"
)

type Config struct {
	Port            string        `yaml:"port"`
	LogLevel        string        `yaml:"log_level"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	Postgres        *Postgres     `yaml:"postgres"`
	Kafka           *Kafka        `yaml:"kafka"`
}

type Postgres struct {
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	DB       string `yaml:"db"`
	User     string `yaml:"user"`
	Password string
}

type Kafka struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

func (cfg *Config) Validate() error {
	if cfg.Postgres.Password == "" {
		return errors.New("postgres password is required")
	}
	return nil
}

func Default() *Config {
	return &Config{
		LogLevel:        "info",
		Port:            ":50054",
		ShutdownTimeout: 10 * time.Second,
		Postgres: &Postgres{
			Port: "5432",
			Host: "postgres",
			DB:   "chatdb",
			User: "chat",
		},
		Kafka: &Kafka{
			Brokers: []string{"kafka:9092"},
			Topic:   "messages",
		},
	}
}

func MustLoad() *Config {
	cfg := Default()
	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	config.MustLoad(cfg)
	return cfg
}
