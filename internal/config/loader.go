package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("CONFIG_PATH environment variable not set")
	}
	if filepath.Ext(configPath) != ".yaml" && filepath.Ext(configPath) != ".yml" {
		panic("config file must be .yaml or .yml")
	}
	file, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}
	cfg := GetDefault()
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config: %v", err))
	}
	fmt.Println(cfg.HTTP.ReadTimeout)
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}
	return cfg
}
