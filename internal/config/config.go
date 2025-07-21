package config

import (
	"errors"
	"time"
)

type Config struct {
	HTTP   HTTP      `yaml:"http"`
	DB     DB        `yaml:"db"`
	JWT    JWT       `yaml:"jwt"`
	Logger LogConfig `yaml:"log"`
}

type HTTP struct {
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type JWT struct {
	Secret string        `yaml:"secret"`
	Expire time.Duration `yaml:"expire"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func (c *Config) Validate() error {
	if c.HTTP.Port == "" {
		return errors.New("http port is required")
	}
	if c.DB.Host == "" {
		return errors.New("db host is required")
	}
	if c.DB.Name == "" {
		return errors.New("db name is required")
	}
	if c.JWT.Secret == "" {
		return errors.New("jwt secret is required")
	}
	return nil
}
