package telegram

type Config struct {
	Token  string
	ChatId string
}

func (c Config) Validate() error {
	if c.Token == "" {
		return ErrMissingToken
	}
	if c.ChatId == "" {
		return ErrMissingChatID
	}
	return nil
}
