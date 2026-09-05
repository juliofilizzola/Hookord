package discord

import (
	"github.com/juliofiliizzola/hookord/internal/domain"
)

func New(cfg Config, repo domain.MessageRepository, client DiscordClient) *Integration {
	return &Integration{
		client: client,
		repo:   repo,
		cfg:    cfg,
	}
}

func NewWithToken(cfg Config, repo domain.MessageRepository) (*Integration, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client, err := NewSessionClient(cfg.Token)
	if err != nil {
		return nil, err
	}

	return New(cfg, repo, client), nil
}

func (integration *Integration) Close() error {
	if integration.client != nil {
		return integration.client.Close()
	}
	return nil
}
