package application

import (
	"net/http"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/events/review"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/config"
	"github.com/juliofiliizzola/hookord/internal/integrations"
	"github.com/rs/zerolog/log"
)

type WebhookService struct {
	config        *config.Config
	repo          domain.MessageRepository
	manager       *integrations.Manager
	reviewCounter *review.Counter
}

func NewWebhookService(cfg *config.Config, repo domain.MessageRepository, manager *integrations.Manager) *WebhookService {
	return &WebhookService{
		config:        cfg,
		repo:          repo,
		manager:       manager,
		reviewCounter: review.NewCounter(repo),
	}
}

func (s *WebhookService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if github.WebHookType(r) == "ping" {
		w.WriteHeader(http.StatusOK)
		return
	}

	secret := []byte("your_github_webhook_secret")
	if s.config != nil && s.config.GithubSecret != "" {
		secret = []byte(s.config.GithubSecret)
	}

	payload, err := github.ValidatePayload(r, secret)
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
		totalReviews, totalReviewers, _ := s.reviewCounter.GetReviewStats(ctx, e.GetPullRequest().GetID())
		err = s.manager.HandlePullRequest(ctx, &integrations.PullRequestEvent{
			Action:         e.GetAction(),
			PullRequest:    e.GetPullRequest(),
			Sender:         e.GetSender(),
			Repository:     e.GetRepo(),
			TotalReviews:   totalReviews,
			TotalReviewers: totalReviewers,
		})
	case *github.IssuesEvent:
		err = s.manager.HandleIssue(ctx, &integrations.IssueEvent{
			Action:     e.GetAction(),
			Issue:      e.GetIssue(),
			Sender:     e.GetSender(),
			Repository: e.GetRepo(),
		})
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
