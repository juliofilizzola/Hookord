package main

import (
	"github.com/juliofilizzola/Hookord/internal/config"
	"github.com/juliofilizzola/Hookord/internal/core"
	"github.com/juliofilizzola/Hookord/internal/httpserver"
	"github.com/juliofilizzola/Hookord/internal/log"
	"github.com/juliofilizzola/Hookord/internal/providers/discord"
	"github.com/juliofilizzola/Hookord/internal/providers/github"
)

func main() {
	cfg := config.LoadConfig()
	//if err := cfg.ValidUrlProvider(); err != nil {
	//	panic(fmt.Errorf("invalid url provider: %v", err))
	//}

	logLevel := cfg.Logging.Level
	logger := log.NewLogger(logLevel)
	logEntry := logger.GetLogger()

	githubProvider := github.New(logEntry)
	discordClient := discord.NewClient(cfg.Discord.WebhookURL, logEntry)

	dispatcher := core.NewDispatcher([]core.OutputPort{discordClient}, logEntry)

	router := httpserver.NewRouter(cfg, logEntry, dispatcher, githubProvider)

	logEntry.Info().Str("port", cfg.HTTP.Port).Msg("🚀 starting server")
	if err := httpserver.Run(router, cfg.HTTP.Port); err != nil {
		logEntry.Fatal().Err(err).Msg("server failed")
	}
	logEntry.Info().Msg("Hookord service started")
}
