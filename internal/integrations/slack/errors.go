package slack

import "errors"

var (
	ErrMissingToken      = errors.New("slack: token is required")
	ErrMissingChannel    = errors.New("slack: channel id is required")
	ErrSessionConnect    = errors.New("slack: err in connect slack api")
	ErrEventNil          = errors.New("slack: event is required")
	ErrClientUnavailable = errors.New("slack: client is unavailable")
)
