package slack

import (
	"context"
	"fmt"

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

	channelId := integration.cfg.ChannelId

	entityID := fmt.Sprintf("%d-%s", event.PullRequest.GetID(), integration.Name())

	mapping, err := integration.repo.GetMapping(ctx, entityID)

	if err != nil {
		return err
	}

	data := BuildPullRequestAttachment(event)

	if mapping == nil || mapping.SlackMessageID == "" {
		channelId, timestamp, err := integration.client.PostMessage(channelId, data)
		if err != nil {
			return err
		}

		if mapping == nil {
			mapping = &domain.MessageMapping{
				EntityID:        entityID,
				SlackMessageID:  channelId,
				SlackTimestamp:  timestamp,
				IntegrationName: integration.Name(),
			}
			err = integration.repo.SaveMapping(ctx, mapping)
			if err != nil {
				return err
			}
		}

		mapping.LastStatus = event.PullRequest.GetState()
		mapping.TotalReviews = event.TotalReviews
		mapping.TotalReviewers = event.TotalReviewers

		return integration.repo.SaveMapping(ctx, mapping)
	}

	targetChannel := mapping.SlackMessageID
	timestampSlack := mapping.SlackTimestamp

	if targetChannel == "" || timestampSlack == "" {
		return ErrEventNotFound
	}
	_, newTimestamp, _, err := integration.client.UpdateMessage(targetChannel, timestampSlack, data)

	if err != nil {
		return err
	}

	mapping.SlackTimestamp = newTimestamp

	return integration.repo.SaveMapping(ctx, mapping)
}
