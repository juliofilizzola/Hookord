package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) Send(channelID string, data *discordgo.MessageSend) (integrations.MessageRef, error) {
	msg, err := integration.client.ChannelMessageSendComplex(channelID, data)
	if err != nil {
		return integrations.MessageRef{}, err
	}
	return integrations.MessageRef{ChannelID: channelID, MessageID: msg.ID}, nil
}

func (integration *Integration) Edit(ref integrations.MessageRef, data *discordgo.MessageSend) (integrations.MessageRef, error) {
	if _, err := integration.client.ChannelMessageEditEmbed(ref.ChannelID, ref.MessageID, data.Embed); err != nil {
		return integrations.MessageRef{}, err
	}
	return ref, nil
}
