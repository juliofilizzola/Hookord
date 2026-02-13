package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/juliofilizzola/Hookord/internal/config"
	"github.com/juliofilizzola/Hookord/internal/core"
	"github.com/juliofilizzola/Hookord/internal/providers/github"
	"github.com/rs/zerolog"
)

func NewRouter(
	cfg *config.Config,
	logger *zerolog.Logger,
	dispatcher *core.Dispatcher,
	githubProvider *github.Provider,
) *gin.Engine {
	router := gin.New()

	router.Use(LoggingMiddleware(logger))

	router.Use(RecoveryMiddleware(logger))

	router.GET("/health", HealthCheck(logger))
	router.GET("/metrics", Metrics(logger))

	handler := NewGithubHandler(cfg, logger, dispatcher, githubProvider)

	router.POST("/github", handler.handle)

	logger.Info().
		Int("routes", len(router.Routes())).
		Str("port", cfg.HTTPPort).Timestamp().
		Msg("Router initialized")
	return router
}

func Run(router *gin.Engine, port string) error {
	return router.Run(":", port)
}
