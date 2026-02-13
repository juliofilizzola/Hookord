package httpserver

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/juliofilizzola/Hookord/internal/config"
	"github.com/juliofilizzola/Hookord/internal/core"
	"github.com/juliofilizzola/Hookord/internal/providers/github"
)

type GithubHandler struct {
	cfg            *config.Config
	logger         zerolog.Logger
	dispatcher     *core.Dispatcher
	githubProvider *github.Provider
}

func NewGithubHandler(cfg *config.Config, logger *zerolog.Logger, dispatcher *core.Dispatcher, githubProvider *github.Provider) *GithubHandler {
	return &GithubHandler{
		cfg:            cfg,
		logger:         logger.With().Str("Componet", "github_handler").Logger(),
		dispatcher:     dispatcher,
		githubProvider: githubProvider,
	}
}

func (handler *GithubHandler) handle(context *gin.Context) {
	requestId := context.GetString("request_id")
	logger := handler.logger.With().Str("request_id", requestId).Logger()

	payload, err := io.ReadAll(context.Request.Body)

	if err != nil {
		logger.
			Error().
			Err(err).
			Msg("Failed to read request body")
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	eventType := context.GetHeader("X-GitHub-Event")
	evt, err := handler.githubProvider.Parse(eventType, payload)
	if err != nil {
		logger.
			Error().
			Err(err).
			Str("event_type", eventType).
			Msg("Failed to parse GitHub event")
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid GitHub event"})
		return
	}

	handler.dispatcher.Dispatch(evt)

	context.JSON(http.StatusOK, gin.H{"status": "event received"})
}
