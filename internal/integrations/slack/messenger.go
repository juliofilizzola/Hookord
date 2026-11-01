package slack

import (
	"github.com/juliofiliizzola/hookord/internal/integrations"
	"github.com/slack-go/slack"
)

func (integration *Integration) Send(channelID string, data slack.Attachment) (integrations.MessageRef, error) {
	channel, timestamp, err := integration.client.PostMessage(channelID, data)
	if err != nil {
		return integrations.MessageRef{}, err
	}
	return integrations.MessageRef{ChannelID: channel, MessageID: timestamp}, nil
}

func (integration *Integration) Edit(ref integrations.MessageRef, data slack.Attachment) (integrations.MessageRef, error) {
	channel, timestamp, _, err := integration.client.UpdateMessage(ref.ChannelID, ref.MessageID, data)
	if err != nil {
		return integrations.MessageRef{}, err
	}
	return integrations.MessageRef{ChannelID: channel, MessageID: timestamp}, nil
}
