package telegram

import "github.com/juliofiliizzola/hookord/internal/domain"

const Name = "Telegram"

func New(cfg Config, repo domain.MessageRepository, client TelegramClient) *Integration {
	return &Integration{
		cfg:    cfg,
		repo:   repo,
		client: client,
	}
}

func NewWithToken(cfg Config, repo domain.MessageRepository) (*Integration, error) {
	client, err := NewSessionClient(cfg.Token)
	if err != nil {
		return nil, err
	}
	return New(cfg, repo, client), nil
}

func (integration *Integration) Name() string {
	return Name
}

func (integration *Integration) Close() error {
	if integration.client != nil {
		return integration.client.Close()
	}
	return nil
}
