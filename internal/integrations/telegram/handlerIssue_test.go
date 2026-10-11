package telegram

import (
	"context"
	"errors"
	"testing"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

func TestHandleIssue(t *testing.T) {
	ctx := context.Background()

	t.Run("Event or Issue nil", func(t *testing.T) {
		integration := &Integration{}
		err := integration.HandleIssue(ctx, nil)
		if err != ErrEventNil {
			t.Errorf("Expected ErrEventNil, got %v", err)
		}
	})

	t.Run("Client nil", func(t *testing.T) {
		integration := &Integration{client: nil}
		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != ErrEventNil {
			t.Errorf("Expected ErrEventNil (client), got %v", err)
		}
	})

	t.Run("Get mapping error", func(t *testing.T) {
		mockErr := errors.New("db error")
		repo := &mockRepo{getErr: mockErr}
		client := &mockTelegramClient{}
		integration := &Integration{client: client, repo: repo}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != mockErr {
			t.Errorf("Expected %v, got %v", mockErr, err)
		}
	})

	t.Run("PostMessage success", func(t *testing.T) {
		repo := &mockRepo{}
		client := &mockTelegramClient{sendId: 999}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChatId: "-12345"}}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.sendCalls != 1 {
			t.Errorf("Expected 1 SendMessage call, got %d", client.sendCalls)
		}
		if repo.saved == nil {
			t.Fatalf("Expected mapping to be saved")
		}
		if repo.saved.SlackMessageID != "-12345" || repo.saved.SlackTimestamp != "999" {
			t.Errorf("Saved mapping mismatch: %+v", repo.saved)
		}
	})

	t.Run("EditMessage success", func(t *testing.T) {
		repo := &mockRepo{
			mappings: map[string]*domain.MessageMapping{
				"1-Telegram": {
					EntityID:        "1-Telegram",
					SlackMessageID:  "-123",
					SlackTimestamp:  "999",
					IntegrationName: "Telegram",
				},
			},
		}
		client := &mockTelegramClient{sendId: 999}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChatId: "-123"}}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.editCalls != 1 {
			t.Errorf("Expected 1 EditMessage call, got %d", client.editCalls)
		}
	})
}
