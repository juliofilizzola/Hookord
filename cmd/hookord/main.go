package main

import (
	"fmt"

	"github.com/juliofilizzola/Hookord/internal/config"
	"github.com/juliofilizzola/Hookord/internal/log"
)

func main() {
	cfg := config.LoadConfig()
	if err := cfg.ValidUrlProvider(); err != nil {
		panic(fmt.Errorf("invalid url provider: %v", err))
	}

	logLevel := cfg.LogLevel
	logger := log.NewLogger(logLevel)
	logEntry := logger.GetLogger()

	logEntry.Info().Msg("Hookord service started")
}
