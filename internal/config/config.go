package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

const DefaultBaseURL = "https://api.thirdear.live"

type Config struct {
	Auth AuthConfig `mapstructure:"auth"`
	API  APIConfig  `mapstructure:"api"`
}

type AuthConfig struct {
	RefreshToken  string    `mapstructure:"refresh_token"`
	IDToken       string    `mapstructure:"id_token"`
	IDTokenExpiry time.Time `mapstructure:"id_token_expiry"`
}

type APIConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(ConfigFile())
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("TWINMIND")
	viper.AutomaticEnv()

	viper.SetDefault("api.base_url", DefaultBaseURL)

	cfg := &Config{}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(*os.PathError); ok {
			cfg.API.BaseURL = DefaultBaseURL
			return cfg, nil
		}
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			cfg.API.BaseURL = DefaultBaseURL
			return cfg, nil
		}
		return nil, err
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(ConfigDir(), 0700); err != nil {
		return err
	}

	viper.Set("auth.refresh_token", cfg.Auth.RefreshToken)
	viper.Set("auth.id_token", cfg.Auth.IDToken)
	viper.Set("auth.id_token_expiry", cfg.Auth.IDTokenExpiry)
	viper.Set("api.base_url", cfg.API.BaseURL)

	viper.SetConfigFile(ConfigFile())
	viper.SetConfigType("yaml")

	if err := viper.WriteConfig(); err != nil {
		return viper.SafeWriteConfig()
	}

	return os.Chmod(ConfigFile(), 0600)
}
