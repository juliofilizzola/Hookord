package telegram

import (
	"github.com/juliofiliizzola/hookord/internal/domain"
)

type Integration struct {
	client TelegramClient
	repo   domain.MessageRepository
	cfg    Config
}
