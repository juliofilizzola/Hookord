package slack

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

	entry := integrations.NewIssueEntry(integration.Name(), integration.cfg.ChannelId, event)
	return integrations.Publish(ctx, integration.repo, integration, entry, BuildIssueAttachment(event))
}
