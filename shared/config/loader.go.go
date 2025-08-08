package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config interface {
	Validate() error
}

func MustLoad(cfg Config) {
	godotenv.Load("./.env")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}
	if filepath.Ext(configPath) != ".yaml" && filepath.Ext(configPath) != ".yml" {
		panic("config file must be .yaml or .yml")
	}
	file, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config: %v", err))
	}
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}
}
