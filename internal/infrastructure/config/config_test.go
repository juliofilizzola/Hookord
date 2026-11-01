package config

import (
	"errors"
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	t.Run("returns value when environment variable is set", func(t *testing.T) {
		t.Setenv("TEST_KEY", "custom_value")

		if got := getEnv("TEST_KEY", "fallback_value"); got != "custom_value" {
			t.Errorf("getEnv() = %q, want %q", got, "custom_value")
		}
	})

	t.Run("returns fallback when environment variable is unset", func(t *testing.T) {
		if got := getEnv("NON_EXISTENT_KEY_12345", "fallback_value"); got != "fallback_value" {
			t.Errorf("getEnv() = %q, want %q", got, "fallback_value")
		}
	})

	t.Run("returns empty string when environment variable is set to empty", func(t *testing.T) {
		t.Setenv("TEST_EMPTY_KEY", "")

		if got := getEnv("TEST_EMPTY_KEY", "fallback_value"); got != "" {
			t.Errorf("getEnv() = %q, want %q", got, "")
		}
	})
}

func TestLoad_Defaults(t *testing.T) {
	// Ensure default variables are unset
	for _, env := range []string{"PORT", "LOG_LEVEL", "APP_ENV"} {
		if val, ok := os.LookupEnv(env); ok {
			defer os.Setenv(env, val)
			os.Unsetenv(env)
		}
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.Port != "8080" {
		t.Errorf("cfg.Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("cfg.LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.Environment != "development" {
		t.Errorf("cfg.Environment = %q, want %q", cfg.Environment, "development")
	}
}

func TestLoad_CustomValuesAndChannelMappings(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "discord-secret-token")
	t.Setenv("GITHUB_SECRET", "github-webhook-secret")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("PORT", "3000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("APP_ENV", "production")

	t.Setenv("DISCORD_CHANNEL_PULL_REQUESTS", "channel_101")
	t.Setenv("DISCORD_CHANNEL_ISSUES", "channel_102")
	t.Setenv("DISCORD_CHANNEL_WORKFLOWS", "channel_103")
	t.Setenv("DISCORD_CHANNEL_REPOSITORY", "channel_104")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.DiscordToken != "discord-secret-token" {
		t.Errorf("cfg.DiscordToken = %q, want %q", cfg.DiscordToken, "discord-secret-token")
	}
	if cfg.GithubSecret != "github-webhook-secret" {
		t.Errorf("cfg.GithubSecret = %q, want %q", cfg.GithubSecret, "github-webhook-secret")
	}
	if cfg.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("cfg.RedisURL = %q, want %q", cfg.RedisURL, "redis://localhost:6379/0")
	}
	if cfg.Port != "3000" {
		t.Errorf("cfg.Port = %q, want %q", cfg.Port, "3000")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("cfg.LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.Environment != "production" {
		t.Errorf("cfg.Environment = %q, want %q", cfg.Environment, "production")
	}

	expectedChannels := map[string]string{
		"pull_requests": "channel_101",
		"issues":        "channel_102",
		"workflows":     "channel_103",
		"repository":    "channel_104",
	}

	for category, expectedChannel := range expectedChannels {
		if got := cfg.ChannelMappings[category]; got != expectedChannel {
			t.Errorf("cfg.ChannelMappings[%q] = %q, want %q", category, got, expectedChannel)
		}
	}
}

func TestValidate(t *testing.T) {
	if err := (&Config{}).Validate(); !errors.Is(err, ErrMissingGithubSecret) || !errors.Is(err, ErrMissingRedisURL) {
		t.Errorf("Validate() = %v, want missing secret and redis url", err)
	}
	if err := (&Config{GithubSecret: "s", RedisURL: "redis://x"}).Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestLoad_SecurityDefaultsAndOverrides(t *testing.T) {
	t.Setenv("RATE_LIMIT_RPS", "")
	t.Setenv("RATE_LIMIT_BURST", "invalid")
	cfg, _ := Load()
	if cfg.RateLimitRPS != 10 || cfg.RateLimitBurst != 20 || cfg.RateLimitTrustProxy {
		t.Errorf("unexpected defaults: rps=%v burst=%d trust=%v", cfg.RateLimitRPS, cfg.RateLimitBurst, cfg.RateLimitTrustProxy)
	}

	t.Setenv("API_KEY", "key")
	t.Setenv("RATE_LIMIT_RPS", "2.5")
	t.Setenv("RATE_LIMIT_BURST", "5")
	t.Setenv("RATE_LIMIT_TRUST_PROXY", "true")
	cfg, _ = Load()
	if cfg.APIKey != "key" || cfg.RateLimitRPS != 2.5 || cfg.RateLimitBurst != 5 || !cfg.RateLimitTrustProxy {
		t.Errorf("unexpected overrides: %+v", cfg)
	}
}

func TestIntegrationEnabled(t *testing.T) {
	cfg := &Config{TelegramToken: "t"}
	if cfg.DiscordEnabled() || cfg.SlackEnabled() || !cfg.TelegramEnabled() {
		t.Errorf("only telegram should be enabled")
	}
}
