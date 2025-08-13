package config

import "time"

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

func DefaultHTTP() HTTP {
	return HTTP{
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		RateLimit:       100,
	}
}
