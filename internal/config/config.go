package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const configPath = "config/config.yaml"

type Config struct {
	Twitch TwitchConfig `yaml:"twitch"`
}

type TwitchConfig struct {
	Username string `yaml:"username" validate:"required"`
	Oauth    string `yaml:"oauth" validate:"required"`
	Channel  string `yaml:"channel" validate:"required"`
}

func NewConfig() (Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func (c Config) validate() error {
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		return err
	}

	return nil
}
