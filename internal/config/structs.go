package config

type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	HTTP struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"http"`
	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
	GitHub struct {
		WebhookSecret string   `mapstructure:"webhook_secret"`
		AllowedRepos  []string `mapstructure:"allowed_repos"`
	} `mapstructure:"github"`
	Discord struct {
		WebhookURL string `mapstructure:"webhook_url"`
	} `mapstructure:"discord"`
}
