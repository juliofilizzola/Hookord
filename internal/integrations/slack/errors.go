package slack

import "errors"

var (
	ErrMissingToken   = errors.New("slack: token is required")
	ErrSessionConnect = errors.New("slack: err in connect slack api")
	ErrEventNil       = errors.New("slack: event is required")
)
