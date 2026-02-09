package config

import "os"

type Config struct {
	DiscordWebhookURL string
}

func Load() *Config {
	return &Config{
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
	}
}
