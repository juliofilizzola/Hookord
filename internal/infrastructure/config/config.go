package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken    string
	GithubSecret    string
	ChannelMappings map[string]string
	SlackToken      string
	SlackChannelId  string
	RedisURL        string
	Port            string
	LogLevel        string
	Environment     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DiscordToken:    os.Getenv("DISCORD_TOKEN"),
		GithubSecret:    os.Getenv("GITHUB_SECRET"),
		ChannelMappings: make(map[string]string),
		SlackToken:      os.Getenv("SLACK_TOKEN"),
		SlackChannelId:  os.Getenv("SLACK_CHANNEL_ID"),
		RedisURL:        os.Getenv("REDIS_URL"),
		Port:            getEnv("PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		Environment:     getEnv("APP_ENV", "development"),
	}

	categories := []string{"pull_requests", "issues", "workflows", "repository"}
	for _, cat := range categories {
		envVar := "DISCORD_CHANNEL_" + strings.ToUpper(cat)
		if val := os.Getenv(envVar); val != "" {
			config.ChannelMappings[cat] = val
		}
	}

	return config, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
