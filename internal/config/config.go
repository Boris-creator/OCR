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

type S3Config struct {
	AccessKeyID     string `envconfig:"S3_ACCESS_KEY_ID"     required:"true"`
	SecretAccessKey string `envconfig:"S3_SECRET_ACCESS_KEY" required:"true"`
	Endpoint        string `envconfig:"S3_ENDPOINT"          required:"true"`
	BucketName      string `envconfig:"S3_BUCKET_NAME"       required:"true"`
}

type DBConfig struct {
	Host     string `required:"true"`
	Port     string `required:"true"`
	User     string `required:"true"`
	Password string `required:"true"`
	Name     string `required:"true"`
}

type Config struct {
	Bot     BotConfig
	Mistral MistralConfig
	S3      S3Config
	DB      DBConfig
}

func Load() (*Config, error) {
	cfg := new(Config)

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}

	return cfg, nil
}
