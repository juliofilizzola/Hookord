package telegram

import (
	"context"
	"fmt"
	"strconv"

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

	channelId := integration.cfg.ChatId

	entityID := fmt.Sprintf("%d-%s", event.Issue.GetID(), integration.Name())

	mapping, err := integration.repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	text := BuildIssueMessage(event)

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
			LastStatus:      event.Issue.GetState(),
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

	mapping.SlackTimestamp = strconv.Itoa(msgId)
	mapping.LastStatus = event.Issue.GetState()

	return integration.repo.SaveMapping(ctx, mapping)
}
