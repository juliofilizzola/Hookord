package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken    string
	GithubSecret    string
	RedisURL        string
	Port            string
	LogLevel        string
	ChannelMappings map[string]string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DiscordToken:    os.Getenv("DISCORD_TOKEN"),
		GithubSecret:    os.Getenv("GITHUB_SECRET"),
		RedisURL:        os.Getenv("REDIS_URL"),
		Port:            getEnv("PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ChannelMappings: make(map[string]string),
	}

	// Load channel mappings from env
	// DISCORD_CHANNEL_PULL_REQUESTS
	// DISCORD_CHANNEL_ISSUES
	// DISCORD_CHANNEL_WORKFLOWS
	// DISCORD_CHANNEL_REPOSITORY
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
