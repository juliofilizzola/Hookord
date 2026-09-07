package slack

import (
	"context"

	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

type Integration struct {
	client SlackClient
	repo   domain.MessageRepository
	cfg    Config
}

func (integration *Integration) HandleIssue(ctx context.Context, event *integrations.IssueEvent) error {
	//TODO implement me
	return nil
}
