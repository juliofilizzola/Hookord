package discord

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/domain"
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

	entityID := strconv.FormatInt(event.PullRequest.GetID(), 10)

	mapping, err := integration.repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	if mapping != nil {
		if event.TotalReviews == 0 && mapping.TotalReviews > 0 {
			event.TotalReviews = mapping.TotalReviews
		}
		if event.TotalReviewers == 0 && mapping.TotalReviewers > 0 {
			event.TotalReviewers = mapping.TotalReviewers
		}
	}

	content := BuildPullRequestContent(event.PullRequest)
	embed := BuildPullRequestEmbed(event)

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
				Reviewers:  make(map[string]bool),
			}
		}

		mapping.DiscordMessageID = msg.ID
		mapping.DiscordChannelID = channelID
		mapping.LastStatus = event.PullRequest.GetState()
		mapping.TotalReviews = event.TotalReviews
		mapping.TotalReviewers = event.TotalReviewers

		return integration.repo.SaveMapping(ctx, mapping)
	}

	targetChannel := mapping.DiscordChannelID

	if targetChannel == "" {
		targetChannel = channelID
	}

	if _, err := integration.client.ChannelMessageEditEmbed(targetChannel, mapping.DiscordMessageID, embed); err != nil {
		return err
	}

	mapping.LastStatus = event.PullRequest.GetState()

	return integration.repo.SaveMapping(ctx, mapping)
}
