package config

import (
	"errors"
	"os"
	"time"

	"github.com/P3rCh1/chat-server/shared/config"
)

type Config struct {
	Port            string        `yaml:"port"`
	Secret          string        `yaml:"secret"`
	Expire          time.Duration `yaml:"expire"`
	LogLevel        string        `yaml:"log_level`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func (cfg *Config) Validate() error {
	if cfg.Secret == "" {
		return errors.New("jwt secret is required")
	}
	return nil
}

func Default() *Config {
	return &Config{
		Expire:          24 * time.Hour,
		LogLevel:        "info",
		Port:            ":50051",
		ShutdownTimeout: 10 * time.Second,
	}
}

func MustLoad() *Config {
	cfg := Default()
	cfg.Secret = os.Getenv("JWT_SECRET")
	config.MustLoad(cfg)
	return cfg
}
