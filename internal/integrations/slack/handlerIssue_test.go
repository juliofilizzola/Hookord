package slack

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func newIssueEvent(id int64, title, state string) *integrations.IssueEvent {
	return &integrations.IssueEvent{
		Action: "opened",
		Issue: &github.Issue{
			ID:     github.Int64(id),
			Number: github.Int(1),
			Title:  github.String(title),
			State:  github.String(state),
			User:   &github.User{Login: github.String("user"), AvatarURL: github.String("https://example.com/avatar.png"), Name: github.String("John")},
		},
		Sender: &github.User{Login: github.String("user"), AvatarURL: github.String("https://example.com/avatar.png")},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner:    &github.User{Name: github.String("org"), AvatarURL: github.String("https://example.com/org.png")},
		},
	}
}

func TestHandleIssue(t *testing.T) {
	ctx := context.Background()

	t.Run("Event or Issue nil", func(t *testing.T) {
		integration := &Integration{}
		err := integration.HandleIssue(ctx, nil)
		if err != ErrEventNil {
			t.Errorf("Expected ErrEventNil, got %v", err)
		}

		err = integration.HandleIssue(ctx, &integrations.IssueEvent{})
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
		client := &mockSlackClient{}
		integration := &Integration{client: client, repo: repo}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != mockErr {
			t.Errorf("Expected %v, got %v", mockErr, err)
		}
	})

	t.Run("PostMessage success", func(t *testing.T) {
		repo := &mockRepo{}
		client := &mockSlackClient{postChannel: "C123", postTimestamp: "ts123"}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChannelId: "C123"}}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.postCalls != 1 {
			t.Errorf("Expected 1 PostMessage call, got %d", client.postCalls)
		}
		if repo.saved == nil {
			t.Fatalf("Expected mapping to be saved")
		}
		if repo.saved.SlackMessageID != "C123" || repo.saved.SlackTimestamp != "ts123" {
			t.Errorf("Saved mapping mismatch: %+v", repo.saved)
		}
	})

	t.Run("PostMessage client error", func(t *testing.T) {
		repo := &mockRepo{}
		mockErr := errors.New("slack error")
		client := &mockSlackClient{postErr: mockErr}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChannelId: "C123"}}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != mockErr {
			t.Errorf("Expected %v, got %v", mockErr, err)
		}
	})

	t.Run("UpdateMessage success", func(t *testing.T) {
		repo := &mockRepo{
			mappings: map[string]*domain.MessageMapping{
				"1-Slack": {
					EntityID:        "1-Slack",
					SlackMessageID:  "C123",
					SlackTimestamp:  "ts123",
					IntegrationName: "Slack",
				},
			},
		}
		client := &mockSlackClient{newTimestamp: "ts456"}
		// Using name "Slack" default
		integration := &Integration{client: client, repo: repo, cfg: Config{ChannelId: "C123"}}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if client.updateCalls != 1 {
			t.Errorf("Expected 1 UpdateMessage call, got %d", client.updateCalls)
		}
		if repo.saved == nil || repo.saved.SlackTimestamp != "ts456" {
			t.Errorf("Expected mapping to be saved with new timestamp")
		}
	})

	t.Run("UpdateMessage missing channel or timestamp", func(t *testing.T) {
		repo := &mockRepo{
			mappings: map[string]*domain.MessageMapping{
				"1-Slack": {
					EntityID:        "1-Slack",
					SlackMessageID:  "", // missing
					SlackTimestamp:  "ts123",
					IntegrationName: "Slack",
				},
			},
		}
		client := &mockSlackClient{}
		integration := &Integration{client: client, repo: repo, cfg: Config{ChannelId: "C123"}}

		err := integration.HandleIssue(ctx, newIssueEvent(1, "title", "open"))
		if err != ErrEventNotFound {
			t.Errorf("Expected ErrEventNotFound, got %v", err)
		}
	})
}
