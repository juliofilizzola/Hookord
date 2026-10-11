package telegram

import (
	"context"
	"fmt"
	"strconv"

	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) HandlePullRequest(ctx context.Context, event *integrations.PullRequestEvent) error {
	if event == nil || event.PullRequest == nil {
		return ErrEventNil
	}

	if integration.client == nil {
		return ErrEventNil
	}

	channelId := integration.cfg.ChatId

	entityID := fmt.Sprintf("%d-%s", event.PullRequest.GetID(), integration.Name())

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

	text := BuildPullRequestMessage(event)

	if mapping == nil {
		messageId, err := integration.client.SendMessage(channelId, text)
		if err != nil {
			return err
		}

		mapping = &domain.MessageMapping{
			EntityID:        entityID,
			SlackMessageID:  channelId,
			SlackTimestamp:  strconv.Itoa(messageId),
			IntegrationName: integration.Name(),
			LastStatus:      event.PullRequest.GetState(),
			TotalReviews:    event.TotalReviews,
			TotalReviewers:  event.TotalReviewers,
		}

		return integration.repo.SaveMapping(ctx, mapping)
	}

	targetChannel := mapping.SlackMessageID
	timestampSlack := mapping.SlackTimestamp

	if targetChannel == "" || timestampSlack == "" {
		return ErrEventNotFound
	}

	msgId, err := strconv.Atoi(timestampSlack)
	if err != nil {
		return err
	}

	_, err = integration.client.EditMessage(targetChannel, msgId, text)
	if err != nil {
		return err
	}

	mapping.LastStatus = event.PullRequest.GetState()
	mapping.TotalReviews = event.TotalReviews
	mapping.TotalReviewers = event.TotalReviewers

	return integration.repo.SaveMapping(ctx, mapping)
}
