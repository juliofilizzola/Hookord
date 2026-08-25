package pullrequest

import (
	"context"
	"strconv"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
)

type EventPayload struct {
	Action         string
	PullRequest    *github.PullRequest
	Sender         *github.User
	Repository     *github.Repository
	TotalReviews   int
	TotalReviewers int
}

func Handle(ctx context.Context, payload *EventPayload, repo domain.MessageRepository, discord domain.DiscordProvider, channelID string) error {
	entityID := strconv.FormatInt(payload.PullRequest.GetID(), 10)

	mapping, err := repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	if mapping != nil {
		payload.TotalReviews = mapping.TotalReviews
		payload.TotalReviewers = mapping.TotalReviewers
	}

	content := BuildContent(payload.PullRequest)

	embed := BuildEmbed(payload)

	if mapping == nil {
		msgID, err := discord.SendMessage(ctx, channelID, content, embed)
		if err != nil {
			return err
		}

		return repo.SaveMapping(ctx, &domain.MessageMapping{
			DiscordMessageID: msgID,
			DiscordChannelID: channelID,
			Repository:       payload.Repository.GetFullName(),
			EntityID:         entityID,
			LastStatus:       payload.PullRequest.GetState(),
			TotalReviews:     payload.TotalReviews,
			TotalReviewers:   payload.TotalReviewers,
		})
	}

	return discord.EditMessage(ctx, mapping.DiscordChannelID, mapping.DiscordMessageID, embed)
}
