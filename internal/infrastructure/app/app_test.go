package app

import (
	"errors"
	"testing"

	"github.com/juliofiliizzola/hookord/internal/infrastructure/config"
	"github.com/juliofiliizzola/hookord/internal/integrations/slack"
	"github.com/juliofiliizzola/hookord/internal/integrations/telegram"
)

func TestBuildIntegrations_NoneConfigured(t *testing.T) {
	_, err := buildIntegrations(&config.Config{}, nil)
	if !errors.Is(err, ErrNoIntegrations) {
		t.Errorf("err = %v, want ErrNoIntegrations", err)
	}
}

func TestBuildIntegrations_InvalidConfig(t *testing.T) {
	_, err := buildIntegrations(&config.Config{TelegramToken: "token"}, nil)
	if !errors.Is(err, telegram.ErrMissingChatID) {
		t.Errorf("err = %v, want telegram.ErrMissingChatID", err)
	}
}

func TestBuildIntegrations_OnlyEnabled(t *testing.T) {
	manager, err := buildIntegrations(&config.Config{SlackToken: "xoxb-test", SlackChannelId: "C1"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := manager.Integrations()
	if len(got) != 1 {
		t.Fatalf("integrations = %d, want 1", len(got))
	}
	if _, ok := got[0].(*slack.Integration); !ok {
		t.Errorf("integration = %T, want *slack.Integration", got[0])
	}
}
