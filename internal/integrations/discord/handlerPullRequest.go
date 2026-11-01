package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) HandlePullRequest(ctx context.Context, event *integrations.PullRequestEvent) error {
	if event == nil || event.PullRequest == nil {
		return ErrNilEvent
	}

	if integration.client == nil {
		return ErrClientUnavailable
	}

	channelID := integration.cfg.GetChannel("pull_requests")
	if channelID == "" {
		return fmt.Errorf("%w: pull_requests", ErrChannelNotConfigured)
	}

	msg := &discordgo.MessageSend{
		Content: BuildPullRequestContent(event.PullRequest),
		Embed:   BuildPullRequestEmbed(event),
	}

	entry := integrations.NewPullRequestEntry(integration.Name(), channelID, event)
	return integrations.Publish(ctx, integration.repo, integration, entry, msg)
}
