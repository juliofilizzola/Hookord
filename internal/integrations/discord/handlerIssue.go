package discord

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/domain"
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

	entityID := strconv.FormatInt(event.Issue.GetID(), 10)

	mapping, err := integration.repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	content := BuildIssueContent(event.Issue)
	embed := BuildIssueEmbed(event)

	if mapping == nil || mapping.DiscordMessageID == "" {
		data := &discordgo.MessageSend{
			Content: content,
			Embed:   embed,
		}

		msg, err := integration.client.ChannelMessageSendComplex(channelID, data)
		if err != nil {
			return err
		}

		repoFullName := ""
		if event.Repository != nil {
			repoFullName = event.Repository.GetFullName()
		}

		if mapping == nil {
			mapping = &domain.MessageMapping{
				EntityID:   entityID,
				Repository: repoFullName,
			}
		}

		mapping.DiscordMessageID = msg.ID
		mapping.DiscordChannelID = channelID
		mapping.LastStatus = event.Issue.GetState()

		return integration.repo.SaveMapping(ctx, mapping)
	}

	targetChannel := mapping.DiscordChannelID
	if targetChannel == "" {
		targetChannel = channelID
	}

	if _, err := integration.client.ChannelMessageEditEmbed(targetChannel, mapping.DiscordMessageID, embed); err != nil {
		return err
	}

	mapping.LastStatus = event.Issue.GetState()
	return integration.repo.SaveMapping(ctx, mapping)
}
