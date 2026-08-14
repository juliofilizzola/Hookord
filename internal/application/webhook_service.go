package application

import (
	"net/http"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/events/issues"
	"github.com/juliofiliizzola/hookord/internal/events/pullrequest"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/config"
	"github.com/rs/zerolog/log"
)

type WebhookService struct {
	config  *config.Config
	repo    domain.MessageRepository
	discord domain.DiscordProvider
}

func NewWebhookService(cfg *config.Config, repo domain.MessageRepository, discord domain.DiscordProvider) *WebhookService {
	return &WebhookService{
		config:  cfg,
		repo:    repo,
		discord: discord,
	}
}

func (s *WebhookService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if github.WebHookType(r) == "ping" {
		w.WriteHeader(http.StatusOK)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(s.config.GithubSecret))
	if err != nil {
		log.Error().Err(err).Msg("failed to validate payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	switch e := event.(type) {
	case *github.PullRequestEvent:
		channelID := s.config.ChannelMappings["pull_requests"]
		if channelID == "" {
			log.Warn().Msg("no channel mapping for pull_requests")
			return
		}
		err = pullrequest.Handle(ctx, &pullrequest.EventPayload{
			Action:      e.GetAction(),
			PullRequest: e.GetPullRequest(),
			Sender:      e.GetSender(),
			Repository:  e.GetRepo(),
		}, s.repo, s.discord, channelID)
	case *github.IssuesEvent:
		channelID := s.config.ChannelMappings["issues"]
		if channelID == "" {
			log.Warn().Msg("no channel mapping for issues")
			return
		}
		err = issues.Handle(ctx, &issues.EventPayload{
			Action:     e.GetAction(),
			Issue:      e.GetIssue(),
			Sender:     e.GetSender(),
			Repository: e.GetRepo(),
		}, s.repo, s.discord, channelID)
	// Add other cases here
	default:
		log.Info().Str("type", github.WebHookType(r)).Msg("received unhandled event")
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to handle event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
