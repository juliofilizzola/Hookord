package discord

import "errors"

var (
	// ErrMissingToken is returned when the Discord bot token is missing.
	ErrMissingToken = errors.New("discord: token is required")
)

// Config holds Discord-specific configuration parameters.
type Config struct {
	Token           string
	ChannelMappings map[string]string
}

// Validate checks whether the required Discord configuration values are present.
func (c Config) Validate() error {
	if c.Token == "" {
		return ErrMissingToken
	}
	return nil
}

// GetChannel returns the configured channel ID for a specific event category.
func (c Config) GetChannel(category string) string {
	if c.ChannelMappings == nil {
		return ""
	}
	return c.ChannelMappings[category]
}
