package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type BotConfig struct {
	Token string `envconfig:"BOT_TOKEN" required:"true"`
}

type MistralConfig struct {
	Token string `envconfig:"MISTRAL_API_KEY" required:"true"`
}

type Config struct {
	Bot     BotConfig
	Mistral MistralConfig
}

func Load() (*Config, error) {
	cfg := new(Config)

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}

	return cfg, nil
}
