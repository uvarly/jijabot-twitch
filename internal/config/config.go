package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const configPath = "./config/config.yaml"

type Config struct {
	Database   DatabaseConfig   `yaml:"database"`
	Phrasebook PhrasebookConfig `yaml:"phrasebook"`
	TwitchBot  TwitchBotConfig  `yaml:"twitch_bot"`
	Oauth      OauthConfig      `yaml:"oauth"`
	JijaBot    JijaBotConfig    `yaml:"jija_bot"`
}

type DatabaseConfig struct {
	Path string `yaml:"path" validate:"required"`
}

type PhrasebookConfig struct {
	Path            string   `yaml:"path" validate:"required"`
	RequiredEntries []string `yaml:"required_entries" validate:"required"`
}

type TwitchBotConfig struct {
	Username     string `yaml:"username" validate:"required"`
	Channel      string `yaml:"channel" validate:"required"`
	ClientID     string `yaml:"client_id" validate:"required"`
	ClientSecret string `yaml:"client_secret" validate:"required"`
}

type OauthConfig struct {
	TokenFilePath string `yaml:"token_file_path" validate:"required"`
}

type JijaBotConfig struct {
	Ranks                []RanksEntry   `yaml:"ranks" validate:"required"`
	Commands             CommandsConfig `yaml:"commands" validate:"required"`
	StreamerTwitchUserID string         `yaml:"streamer_twitch_user_id" validate:"required"`
}

type RanksEntry struct {
	Name      string `yaml:"name" validate:"required"`
	MinRating int    `yaml:"min_rating" validate:"min=0"`
}

type CommandsConfig struct {
	Daily DailyConfig `yaml:"daily"`
	Bet   BetConfig   `yaml:"bet"`
	Duel  DuelConfig  `yaml:"duel"`
}

type DailyConfig struct {
	JijaCoinAmount int64 `yaml:"jija_coin_amount" validate:"required"`
	ResetHourUTC   int   `yaml:"reset_hour_utc" validate:"min=0,max=23"`
}

type BetConfig struct {
	BaseProbability float64 `yaml:"base_probability" validate:"required"`
	PayoutMultiple  float64 `yaml:"payout_multiple" validate:"required"`
	DailyLimit      int     `yaml:"daily_limit" validate:"required"`
	ResetHourUTC    int     `yaml:"reset_hour_utc" validate:"min=0,max=23"`
}

type DuelConfig struct {
	MMR MMRConfig

	DailyLimit     int           `yaml:"daily_limit" validate:"required"`
	ResetHourUTC   int           `yaml:"reset_hour_utc" validate:"min=0,max=23"`
	ExpiryDuration time.Duration `yaml:"expiry_duration" validate:"required"`
}

type MMRConfig struct {
	ELO ELOConfig `yaml:"elo"`
}

type ELOConfig struct {
	DefaultRating int `yaml:"default_rating" validate:"required"`
	KFactor       int `yaml:"k_factor" validate:"required"`
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
