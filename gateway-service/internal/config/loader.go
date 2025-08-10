package config

import (
	"time"

	"github.com/P3rCh1/chat-server/gateway-service/shared/config"
	"github.com/P3rCh1/chat-server/gateway-service/shared/logger"
)

type Config struct {
	HTTP     HTTP     `yaml:"http"`
	Services Services `yaml:"services"`
	LogLVL   string   `yaml:"log_level"`
}

type HTTP struct {
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	AllowedOrigins  []string      `yaml:"allowed_origins"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	RateLimit       int           `yaml:"rate_limit"`
}

type Services struct {
	SessionAddr string           `yaml:"session_addr"`
	UserAddr    string           `yaml:"user_addr"`
	RoomsAddr   string           `yaml:"rooms_addr"`
	Timeouts    TimeoutsServices `yaml:"timeouts"`
}

type TimeoutsServices struct {
	Session time.Duration `yaml:"session"`
	User    time.Duration `yaml:"user"`
	Rooms   time.Duration `yaml:"rooms"`
}

func (cfg *Config) Validate() error {
	return nil
}

func Default() *Config {
	return &Config{
		HTTP: HTTP{
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			RateLimit:       100,
		},
		Services: Services{
			SessionAddr: "session:50051",
			UserAddr:    "user:50052",
			Timeouts: TimeoutsServices{
				Session: 1 * time.Second,
				User:    1 * time.Second,
			},
		},
		LogLVL: logger.InfoLVL,
	}
}

func MustLoad() *Config {
	cfg := Default()
	config.MustLoad(cfg)
	return cfg
}
