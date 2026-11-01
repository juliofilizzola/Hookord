package integrations

import (
	"context"
	"fmt"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

type MessageRef struct {
	ChannelID string
	MessageID string
}

type Messenger[T any] interface {
	Send(channelID string, msg T) (MessageRef, error)
	Edit(ref MessageRef, msg T) (MessageRef, error)
}

type Entry struct {
	Integration    string
	EntityID       int64
	ChannelID      string
	Repository     string
	Status         string
	TotalReviews   int
	TotalReviewers int
}

func MappingKey(entityID int64, integration string) string {
	return fmt.Sprintf("%d-%s", entityID, integration)
}

func Publish[T any](ctx context.Context, repo domain.MessageRepository, messenger Messenger[T], entry Entry, msg T) error {
	key := MappingKey(entry.EntityID, entry.Integration)

	mapping, err := repo.GetMapping(ctx, key)
	if err != nil {
		return err
	}

	var ref MessageRef
	if mapping == nil || mapping.MessageID == "" {
		ref, err = messenger.Send(entry.ChannelID, msg)
	} else {
		ref = MessageRef{ChannelID: mapping.ChannelID, MessageID: mapping.MessageID}
		if ref.ChannelID == "" {
			ref.ChannelID = entry.ChannelID
		}
		ref, err = messenger.Edit(ref, msg)
	}
	if err != nil {
		return err
	}

	if mapping == nil {
		mapping = &domain.MessageMapping{
			EntityID:        key,
			IntegrationName: entry.Integration,
			Repository:      entry.Repository,
		}
	}

	mapping.ChannelID = ref.ChannelID
	mapping.MessageID = ref.MessageID
	mapping.LastStatus = entry.Status
	mapping.TotalReviews = entry.TotalReviews
	mapping.TotalReviewers = entry.TotalReviewers

	return repo.SaveMapping(ctx, mapping)
}
