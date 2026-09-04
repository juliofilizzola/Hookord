package discord

import "github.com/juliofiliizzola/hookord/internal/domain"

type Integration struct {
	client DiscordClient
	repo   domain.MessageRepository
	cfg    Config
}
