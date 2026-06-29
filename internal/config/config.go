package config

import (
	"os"
	"strings"
)

func loadEnv() *EnvConfig {
	envs := &EnvConfig{}

	envs.Port = os.Getenv("PORT")
	envs.Level = os.Getenv("LOG_LEVEL")
	envs.WebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	envs.WebhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")

	if allowedRepos := os.Getenv("ALLOWED_REPOS"); allowedRepos != "" {
		envs.AllowedRepos = strings.Split(allowedRepos, ",")

		for indexRepos := range envs.AllowedRepos {
			envs.AllowedRepos[indexRepos] = strings.TrimSpace(envs.AllowedRepos[indexRepos])
		}
	}

	return envs
}

func LoadConfig() *Config {
	env := loadEnv()

	cfg := &Config{}

	cfg.App.Name = "Hookord"
	cfg.App.Version = "1.0.0"
	cfg.HTTP.Port = "8080"
	cfg.Logging.Level = "info"

	if env.Port != "" {
		cfg.HTTP.Port = env.Port
	}
	if env.Level != "" {
		cfg.Logging.Level = env.Level
	}
	if env.WebhookURL != "" {
		cfg.Discord.WebhookURL = env.WebhookURL
	}
	if env.WebhookSecret != "" {
		cfg.GitHub.WebhookSecret = env.WebhookSecret
	}
	if len(env.AllowedRepos) > 0 {
		cfg.GitHub.AllowedRepos = env.AllowedRepos
	}

	return cfg
}
