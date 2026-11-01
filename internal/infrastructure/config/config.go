package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	ErrMissingGithubSecret = errors.New("config: GITHUB_SECRET is required")
	ErrMissingRedisURL     = errors.New("config: REDIS_URL is required")
)

type Config struct {
	DiscordToken    string
	GithubSecret    string
	ChannelMappings map[string]string
	SlackToken      string
	SlackChannelId  string
	TelegramToken   string
	TelegramChatId  string
	RedisURL        string
	Port            string
	LogLevel        string
	Environment     string

	APIKey string

	RateLimitRPS        float64
	RateLimitBurst      int
	RateLimitTrustProxy bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DiscordToken:        os.Getenv("DISCORD_TOKEN"),
		GithubSecret:        os.Getenv("GITHUB_SECRET"),
		ChannelMappings:     make(map[string]string),
		SlackToken:          os.Getenv("SLACK_TOKEN"),
		SlackChannelId:      os.Getenv("SLACK_CHANNEL_ID"),
		TelegramToken:       os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatId:      os.Getenv("TELEGRAM_CHAT_ID"),
		RedisURL:            os.Getenv("REDIS_URL"),
		Port:                getEnv("PORT", "8080"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		Environment:         getEnv("APP_ENV", "development"),
		APIKey:              os.Getenv("API_KEY"),
		RateLimitRPS:        getEnvFloat("RATE_LIMIT_RPS", 10),
		RateLimitBurst:      getEnvInt("RATE_LIMIT_BURST", 20),
		RateLimitTrustProxy: getEnvBool("RATE_LIMIT_TRUST_PROXY", false),
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

func (c *Config) Validate() error {
	var errs []error
	if c.GithubSecret == "" {
		errs = append(errs, ErrMissingGithubSecret)
	}
	if c.RedisURL == "" {
		errs = append(errs, ErrMissingRedisURL)
	}
	return errors.Join(errs...)
}

func (c *Config) DiscordEnabled() bool  { return c.DiscordToken != "" }
func (c *Config) SlackEnabled() bool    { return c.SlackToken != "" }
func (c *Config) TelegramEnabled() bool { return c.TelegramToken != "" }

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if value, err := strconv.ParseFloat(os.Getenv(key), 64); err == nil {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return value
	}
	return fallback
}
