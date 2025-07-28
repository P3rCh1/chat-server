package config

import (
	"errors"
	"time"
)

type Config struct {
	PKG       Package   `yaml:"pkg"`
	HTTP      HTTP      `yaml:"http"`
	WebSocket Websocket `yaml:"websocket"`
	DB        DB        `yaml:"db"`
	JWT       JWT       `yaml:"jwt"`
	Logger    LogConfig `yaml:"log"`
}

func GetDefault() *Config {
	return &Config{
		PKG: Package{
			SystemUsername: "system",
			ErrorUsername:  "error",
		},
		HTTP: HTTP{
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 20 * time.Second,
			RateLimit:       100,
		},
		WebSocket: Websocket{
			WriteBufSize:      4096,
			ReadBufSize:       4096,
			MsgBufSize:        20,
			MsgMaxSize:        16384,
			MsgMaxLength:      2000,
			WriteWait:         10 * time.Second,
			PongWait:          60 * time.Second,
			PingPeriod:        54 * time.Second,
			MaxFailedPings:    3,
			EnableCompression: true,
			CheckOrigin:       false,
		},
		JWT: JWT{
			Expire: 24 * time.Hour,
		},
		Logger: LogConfig{
			Level: "debug",
		},
	}
}

type Package struct {
	SystemUsername string `yaml:"system_username"`
	ErrorUsername  string `yaml:"error_username"`
}

type HTTP struct {
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	RateLimit       int           `yaml:"rate_limit"`
}

type Websocket struct {
	MsgMaxSize        int           `yaml:"msg_max_size"`
	MsgMaxLength      int           `yaml:"msg_max_length"`
	WriteBufSize      int           `yaml:"write_buf_size"`
	ReadBufSize       int           `yaml:"read_buf_size"`
	MsgBufSize        int           `yaml:"msg_buf_size"`
	EnableCompression bool          `yaml:"enable_compression"`
	WriteWait         time.Duration `yaml:"write_wait"`
	PongWait          time.Duration `yaml:"pong_wait"`
	PingPeriod        time.Duration `yaml:"ping_period"`
	MaxFailedPings    int           `yaml:"max_failed_pings"`
	CheckOrigin       bool          `yaml:"check_origin"`
	AllowedOrigins    []string      `yaml:"allowed_origins"`
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
	if c.HTTP.Host == "" {
		return errors.New("http host is required")
	}
	if c.DB.Host == "" {
		return errors.New("db host is required")
	}
	if c.DB.Name == "" {
		return errors.New("db name is required")
	}
	if c.DB.User == "" {
		return errors.New("db user is required")
	}
	if c.DB.Password == "" {
		return errors.New("db password is required")
	}
	if c.JWT.Secret == "" {
		return errors.New("jwt secret is required")
	}
	return nil
}
