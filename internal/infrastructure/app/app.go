package app

import (
	"context"
	"errors"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juliofiliizzola/hookord/internal/application"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/config"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/http"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/logger"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/redis"
	"github.com/juliofiliizzola/hookord/internal/integrations"
	"github.com/juliofiliizzola/hookord/internal/integrations/discord"
	"github.com/juliofiliizzola/hookord/internal/integrations/slack"
	"github.com/juliofiliizzola/hookord/internal/integrations/telegram"
	"github.com/rs/zerolog/log"
)

var ErrNoIntegrations = errors.New("app: no integration configured (set DISCORD_TOKEN, SLACK_TOKEN or TELEGRAM_TOKEN)")

type App struct {
	cfg *config.Config
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
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

	integrationManager, err := buildIntegrations(a.cfg, repo)
	if err != nil {
		return err
	}

	defer func() {
		if err := integrationManager.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close integrations")
		}
	}()

	webhookService := application.NewWebhookService(a.cfg.GithubSecret, repo, integrationManager)
	srv := http.NewServer(http.Config{
		Port:   a.cfg.Port,
		APIKey: a.cfg.APIKey,
		RateLimit: http.RateLimitConfig{
			RequestsPerSecond: a.cfg.RateLimitRPS,
			Burst:             a.cfg.RateLimitBurst,
			TrustProxy:        a.cfg.RateLimitTrustProxy,
		},
	}, nethttp.HandlerFunc(webhookService.HandleWebhook))

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

func buildIntegrations(cfg *config.Config, repo domain.MessageRepository) (*integrations.Manager, error) {
	manager := integrations.NewManager()

	register := func(integration integrations.Integration, err error) error {
		if err != nil {
			_ = manager.Close()
			return err
		}
		manager.Register(integration)
		log.Info().Str("integration", integration.Name()).Msg("integration enabled")
		return nil
	}

	if cfg.DiscordEnabled() {
		if err := register(discord.NewWithToken(discord.Config{
			Token:           cfg.DiscordToken,
			ChannelMappings: cfg.ChannelMappings,
		}, repo)); err != nil {
			return nil, err
		}
	}

	if cfg.SlackEnabled() {
		if err := register(slack.NewWithToken(slack.Config{
			Token:     cfg.SlackToken,
			ChannelId: cfg.SlackChannelId,
		}, repo)); err != nil {
			return nil, err
		}
	}

	if cfg.TelegramEnabled() {
		if err := register(telegram.NewWithToken(telegram.Config{
			Token:  cfg.TelegramToken,
			ChatId: cfg.TelegramChatId,
		}, repo)); err != nil {
			return nil, err
		}
	}

	if len(manager.Integrations()) == 0 {
		return nil, ErrNoIntegrations
	}

	return manager, nil
}
