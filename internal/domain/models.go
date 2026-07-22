package domain

import "context"

type Event struct {
	ID       string
	Source   string // "github"
	Type     string // e.g., "pull_request", "issues"
	Action   string // e.g., "opened", "closed"
	Payload  interface{}
	Metadata map[string]string
}

type MessageMapping struct {
	EventID          string
	DiscordMessageID string
	DiscordChannelID string
	Repository       string
	EntityID         string
	LastStatus       string
}

type MessageRepository interface {
	SaveMapping(ctx context.Context, mapping *MessageMapping) error
	GetMapping(ctx context.Context, entityID string) (*MessageMapping, error)
	DeleteMapping(ctx context.Context, entityID string) error
}

type DiscordProvider interface {
	SendMessage(ctx context.Context, channelID string, embed interface{}) (string, error)
	EditMessage(ctx context.Context, channelID string, messageID string, embed interface{}) error
}
