package telegram

import (
	"strconv"

	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) Send(chatID string, text string) (integrations.MessageRef, error) {
	messageID, err := integration.client.SendMessage(chatID, text)
	if err != nil {
		return integrations.MessageRef{}, err
	}
	return integrations.MessageRef{ChannelID: chatID, MessageID: strconv.Itoa(messageID)}, nil
}

func (integration *Integration) Edit(ref integrations.MessageRef, text string) (integrations.MessageRef, error) {
	messageID, err := strconv.Atoi(ref.MessageID)
	if err != nil {
		return integrations.MessageRef{}, err
	}
	if _, err := integration.client.EditMessage(ref.ChannelID, messageID, text); err != nil {
		return integrations.MessageRef{}, err
	}
	return ref, nil
}
