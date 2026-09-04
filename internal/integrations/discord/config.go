package discord

import "errors"

var (
	ErrMissingToken = errors.New("discord: token is required")
)

type Config struct {
	Token           string
	ChannelMappings map[string]string
}

func (c Config) Validate() error {
	if c.Token == "" {
		return ErrMissingToken
	}
	return nil
}

func (c Config) GetChannel(category string) string {
	if c.ChannelMappings == nil {
		return ""
	}
	return c.ChannelMappings[category]
}
