package slack

import (
	"context"
	"fmt"

	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) HandleIssue(ctx context.Context, event *integrations.IssueEvent) error {
	if event == nil || event.Issue == nil {
		return ErrEventNil
	}

	if integration.client == nil {
		return ErrEventNil
	}

	channelId := integration.cfg.ChannelId

	entityID := fmt.Sprintf("%d-%s", event.Issue.GetID(), integration.Name())

	mapping, err := integration.repo.GetMapping(ctx, entityID)

	if err != nil {
		return err
	}

	data := BuildIssueAttachment(event)

	if mapping == nil {
		channelId, timestamp, err := integration.client.PostMessage(channelId, data)
		if err != nil {
			return err
		}

		mapping = &domain.MessageMapping{
			EntityID:        entityID,
			SlackMessageID:  channelId,
			SlackTimestamp:  timestamp,
			IntegrationName: integration.Name(),
			LastStatus:      event.Issue.GetState(),
		}

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
	mapping.LastStatus = event.Issue.GetState()

	return integration.repo.SaveMapping(ctx, mapping)
}
