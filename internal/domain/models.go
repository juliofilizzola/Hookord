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
	EventID          string          `json:"event_id,omitempty"`
	DiscordMessageID string          `json:"discord_message_id,omitempty"`
	DiscordChannelID string          `json:"discord_channel_id,omitempty"`
	Repository       string          `json:"repository,omitempty"`
	EntityID         string          `json:"entity_id,omitempty"`
	LastStatus       string          `json:"last_status,omitempty"`
	TotalReviews     int             `json:"total_reviews"`
	TotalReviewers   int             `json:"total_reviewers"`
	Reviewers        map[string]bool `json:"reviewers,omitempty"`
}

type MessageRepository interface {
	SaveMapping(ctx context.Context, mapping *MessageMapping) error
	GetMapping(ctx context.Context, entityID string) (*MessageMapping, error)
	DeleteMapping(ctx context.Context, entityID string) error
}

type DiscordProvider interface {
	SendMessage(ctx context.Context, channelID string, content string, embed interface{}) (string, error)
	EditMessage(ctx context.Context, channelID string, messageID string, embed interface{}) error
}
