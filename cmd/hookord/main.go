package main

import (
	"github.com/juliofiliizzola/hookord/internal/infrastructure/app"
	"github.com/rs/zerolog/log"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	if err := application.Run(); err != nil {
		log.Fatal().Err(err).Msg("application finished with error")
	}
}
