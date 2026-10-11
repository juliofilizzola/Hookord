package telegram

import "errors"

var (
	ErrEventNil       = errors.New("telegram: event is required")
	ErrSessionConnect = errors.New("telegram: error to connect session")
	ErrEventNotFound  = errors.New("telegram: event not found in mappings")
)
