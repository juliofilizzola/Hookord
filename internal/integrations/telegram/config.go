package telegram

import "errors"

type Config struct {
	Token  string
	ChatId string
}

func (c Config) Validate() error {
	if c.Token == "" {
		return errors.New("token is required")
	}
	return nil
}
