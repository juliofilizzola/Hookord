package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/domain"
)

type provider struct {
	session *discordgo.Session
}

func NewProvider(token string) (domain.DiscordProvider, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to Discord")

	err = dg.Open()
	if err != nil {
		return nil, err
	}

	defer func(dg *discordgo.Session) {
		err := dg.Close()
		if err != nil {
			fmt.Println("Failed to close discord session")
		}
	}(dg)

	return &provider{session: dg}, nil
}

func (p *provider) SendMessage(ctx context.Context, channelID string, embed interface{}) (string, error) {
	e, ok := embed.(*discordgo.MessageEmbed)
	if !ok {
		return "", fmt.Errorf("invalid embed type")
	}

	msg, err := p.session.ChannelMessageSendEmbed(channelID, e)
	if err != nil {
		return "", err
	}

	return msg.ID, nil
}

func (p *provider) EditMessage(ctx context.Context, channelID string, messageID string, embed interface{}) error {
	e, ok := embed.(*discordgo.MessageEmbed)
	if !ok {
		return fmt.Errorf("invalid embed type")
	}

	_, err := p.session.ChannelMessageEditEmbed(channelID, messageID, e)
	return err
}
