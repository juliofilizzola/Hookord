package slack

type Config struct {
	Token     string
	ChannelId string
}

func (c Config) Validate() error {
	if c.Token == "" {
		return ErrMissingToken
	}
	if c.ChannelId == "" {
		return ErrMissingChannel
	}
	return nil
}
