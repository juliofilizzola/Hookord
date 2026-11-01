package telegram

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

	entry := integrations.NewPullRequestEntry(integration.Name(), integration.cfg.ChatId, event)
	return integrations.Publish(ctx, integration.repo, integration, entry, BuildPullRequestMessage(event))
}
