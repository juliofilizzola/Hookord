package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juliofiliizzola/hookord/internal/application"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/config"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/discord"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/http"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/logger"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/redis"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &App{cfg: cfg}, nil
}

func (a *App) Run() error {
	logger.Setup(logger.Config{
		Level:       a.cfg.LogLevel,
		Environment: a.cfg.Environment,
	})

	repo, err := redis.NewRepository(a.cfg.RedisURL)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to redis")
		return err
	}

	discordProvider, err := discord.NewProvider(a.cfg.DiscordToken)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to discord")
		return err
	}

	webhookService := application.NewWebhookService(a.cfg, repo, discordProvider)
	srv := http.NewServer(a.cfg.Port, webhookService)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
		return err
	}

	log.Info().Msg("server exited")
	return nil
}
