package slack

import (
	"github.com/slack-go/slack"
)

type SlackClient interface {
	PostMessage(channelId string, data slack.Attachment) (string, string, error)
	UpdateMessage(channelId, ts string, data slack.Attachment) (string, string, string, error)
}

type sessionClientSlack struct {
	session *slack.Client
}

func NewSessionClient(token string) (SlackClient, error) {
	session := slack.New(token, slack.OptionDebug(true))

	if session == nil {
		return nil, ErrSessionConnect
	}

	return &sessionClientSlack{
		session: session,
	}, nil
}

func (sl *sessionClientSlack) PostMessage(channelId string, data slack.Attachment) (string, string, error) {
	return sl.session.PostMessage(channelId, slack.MsgOptionAttachments(data))
}

func (sl *sessionClientSlack) UpdateMessage(channelId, ts string, data slack.Attachment) (string, string, string, error) {
	return sl.session.UpdateMessage(channelId, ts, slack.MsgOptionAttachments(data))
}
