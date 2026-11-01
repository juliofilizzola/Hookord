package telegram

import "errors"

var (
	ErrMissingToken      = errors.New("telegram: token is required")
	ErrMissingChatID     = errors.New("telegram: chat id is required")
	ErrEventNil          = errors.New("telegram: event is required")
	ErrClientUnavailable = errors.New("telegram: client is unavailable")
	ErrSessionConnect    = errors.New("telegram: error to connect session")
)
