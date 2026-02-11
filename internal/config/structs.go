package config

type Config struct {
	HTTPPort          string   `mapstructure:"http_port"`
	GitHubSecret      string   `mapstructure:"github_webhook_secret"`
	AllowedRepo       []string `mapstructure:"github_allowed_repos"`
	DiscordWebhookURL string   `mapstructure:"discord_webhook_url"`
	LogLevel          string   `mapstructure:"log_level"`
	AppName           string   `mapstructure:"app_name"`
	AppVersion        string   `mapstructure:"app_version"`
	RequestId         string   `mapstructure:"request_id"`
}
