package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) HandleIssue(ctx context.Context, event *integrations.IssueEvent) error {
	if event == nil || event.Issue == nil {
		return ErrNilEvent
	}

	if integration.client == nil {
		return ErrClientUnavailable
	}

	channelID := integration.cfg.GetChannel("issues")
	if channelID == "" {
		return fmt.Errorf("%w: issues", ErrChannelNotConfigured)
	}

	msg := &discordgo.MessageSend{
		Content: BuildIssueContent(event.Issue),
		Embed:   BuildIssueEmbed(event),
	}

	entry := integrations.NewIssueEntry(integration.Name(), channelID, event)
	return integrations.Publish(ctx, integration.repo, integration, entry, msg)
}
