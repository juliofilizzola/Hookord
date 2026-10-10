package slack

import (
	"github.com/juliofiliizzola/hookord/internal/domain"
)

type Integration struct {
	client SlackClient
	repo   domain.MessageRepository
	cfg    Config
}
