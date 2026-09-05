package discord

import "errors"

var (
	ErrClientUnavailable = errors.New("discord: client is unavailable")

	ErrChannelNotConfigured = errors.New("discord: channel not configured")

	ErrNilEvent = errors.New("discord: event payload is nil")
)
