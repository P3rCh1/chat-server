package config

import (
	"github.com/P3rCh1/chat-server/gateway-service/pkg/config"
	"github.com/P3rCh1/chat-server/gateway-service/pkg/logger"
)

type Config struct {
	HTTP      HTTP      `yaml:"http"`
	Websocket Websocket `yaml:"websocket"`
	Services  Services  `yaml:"services"`
	LogLVL    string    `yaml:"log_level"`
	Kafka     Kafka     `yaml:"kafka"`
}

func (cfg *Config) Validate() error {
	return nil
}

func Default() *Config {
	return &Config{
		HTTP:      DefaultHTTP(),
		Websocket: DefaultWebsocket(),
		Services:  DefaultServices(),
		Kafka:     DefaultKafka(),
		LogLVL:    logger.InfoLVL,
	}
}

func MustLoad() *Config {
	cfg := Default()
	config.MustLoad(cfg)
	return cfg
}
