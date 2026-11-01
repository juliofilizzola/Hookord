package slack

import (
	"context"

	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) HandlePullRequest(ctx context.Context, event *integrations.PullRequestEvent) error {
	if event == nil || event.PullRequest == nil {
		return ErrEventNil
	}

	if integration.client == nil {
		return ErrClientUnavailable
	}

	entry := integrations.NewPullRequestEntry(integration.Name(), integration.cfg.ChannelId, event)
	return integrations.Publish(ctx, integration.repo, integration, entry, BuildPullRequestAttachment(event))
}
