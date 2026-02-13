package config

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("$HOME/.github.com/juliofilizzola/Hookord")
	viper.AddConfigPath("/etc/gihub.com/juliofilizzola/Hookord")
	viper.AutomaticEnv()

	viper.SetEnvPrefix("hookord")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("request_id", uuid.NewString())
	viper.SetDefault("app_version", "0.0.1")
	viper.SetDefault("app_name", "hookord")
	viper.SetDefault("port", "8080")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file, using defaults", err)
	}
	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}
	fmt.Printf("Config loaded: %+v\n", cfg)

	if cfg.GitHub.WebhookSecret == "" {
		panic("GitHub secret is required")
	}

	if cfg.Discord.WebhookURL == "" {
		panic("Discord webhook URL is required")
	}

	return cfg
}

func (c *Config) ValidUrlProvider() error {
	if !strings.HasPrefix(c.GitHub.WebhookSecret, "https://") || !strings.HasPrefix(c.Discord.WebhookURL, "/slack") {
		return fmt.Errorf("invalid url provider")
	}
	return nil
}
