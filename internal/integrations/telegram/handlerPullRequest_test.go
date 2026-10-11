package telegram

import (
	"context"
	"errors"
	"testing"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

func TestHandlePullRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("Event or PR nil", func(t *testing.T) {
		integration := &Integration{}
		err := integration.HandlePullRequest(ctx, nil)
		if err != ErrEventNil {
			t.Errorf("Expected ErrEventNil, got %v", err)
		}
	})

	t.Run("Client nil", func(t *testing.T) {
		integration := &Integration{client: nil}
		err := integration.HandlePullRequest(ctx, newPREvent(1, "title", "open"))
		if err != ErrEventNil {
			t.Errorf("Expected ErrEventNil (client), got %v", err)
		}
	})

	t.Run("Get mapping error", func(t *testing.T) {
		mockErr := errors.New("db error")
		repo := &mockRepo{getErr: mockErr}
		client := &mockTelegramClient{}
		integration := &Integration{client: client, repo: repo}

		err := integration.HandlePullRequest(ctx, newPREvent(1, "title", "open"))
		if err != mockErr {
			t.Errorf("Expected %v, got %v", mockErr, err)
		}
	})

	t.Run("PostMessage success", func(t *testing.T) {
		repo := &mockRepo{}
		client := &mockTelegramClient{sendId: 12345}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChatId: "-12345"}}

		err := integration.HandlePullRequest(ctx, newPREvent(1, "title", "open"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.sendCalls != 1 {
			t.Errorf("Expected 1 SendMessage call, got %d", client.sendCalls)
		}
		if repo.saved == nil {
			t.Fatalf("Expected mapping to be saved")
		}
		if repo.saved.SlackMessageID != "-12345" || repo.saved.SlackTimestamp != "12345" {
			t.Errorf("Saved mapping mismatch: %+v", repo.saved)
		}
	})

	t.Run("PostMessage client error", func(t *testing.T) {
		repo := &mockRepo{}
		mockErr := errors.New("telegram error")
		client := &mockTelegramClient{sendErr: mockErr}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChatId: "-123"}}

		err := integration.HandlePullRequest(ctx, newPREvent(1, "title", "open"))
		if err != mockErr {
			t.Errorf("Expected %v, got %v", mockErr, err)
		}
	})

	t.Run("EditMessage success", func(t *testing.T) {
		repo := &mockRepo{
			mappings: map[string]*domain.MessageMapping{
				"1-Telegram": {
					EntityID:        "1-Telegram",
					SlackMessageID:  "-123",
					SlackTimestamp:  "12345",
					IntegrationName: "Telegram",
				},
			},
		}
		client := &mockTelegramClient{sendId: 12345}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChatId: "-123"}}

		err := integration.HandlePullRequest(ctx, newPREvent(1, "title", "open"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.editCalls != 1 {
			t.Errorf("Expected 1 EditMessage call, got %d", client.editCalls)
		}
		if repo.saved == nil {
			t.Errorf("Expected mapping to be saved with new stats")
		}
	})
}
