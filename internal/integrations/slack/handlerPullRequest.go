package slack

import (
	"context"
	"fmt"

	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func (integration *Integration) HandlePullRequest(ctx context.Context, event *integrations.PullRequestEvent) error {
	if event == nil || event.PullRequest == nil {
		return ErrEventNil
	}

	if integration.client == nil {
		return ErrEventNil
	}

	channelId := integration.cfg.ChannelId

	//
	//entityID := strconv.FormatInt(event.PullRequest.GetID(), 10) + "-slack"
	//
	//mapping

	data := BuildPullRequestAttachment(event)
	fmt.Println(data)

	msg, timestamp, err := integration.client.PostMessage(channelId, data)

	fmt.Println(timestamp)
	fmt.Println(msg)

	if err != nil {
		return err
	}

	return nil
}
