package telegram

import (
	"context"

	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) HandleIssue(ctx context.Context, event *integrations.IssueEvent) error {
	if event == nil || event.Issue == nil {
		return ErrEventNil
	}

	if integration.client == nil {
		return ErrClientUnavailable
	}

	entry := integrations.NewIssueEntry(integration.Name(), integration.cfg.ChatId, event)
	return integrations.Publish(ctx, integration.repo, integration, entry, BuildIssueMessage(event))
}
