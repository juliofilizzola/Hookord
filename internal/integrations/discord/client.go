package discord

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordClient interface {
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend) (*discordgo.Message, error)
	ChannelMessageEditEmbed(channelID string, messageID string, embed *discordgo.MessageEmbed) (*discordgo.Message, error)
	Close() error
}

type sessionClient struct {
	session *discordgo.Session
}

func NewSessionClient(token string) (DiscordClient, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	if err := dg.Open(); err != nil {
		return nil, err
	}

	return &sessionClient{session: dg}, nil
}

func (c *sessionClient) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend) (*discordgo.Message, error) {
	return c.session.ChannelMessageSendComplex(channelID, data)
}

func (c *sessionClient) ChannelMessageEditEmbed(channelID string, messageID string, embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	return c.session.ChannelMessageEditEmbed(channelID, messageID, embed)
}

func (c *sessionClient) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}
