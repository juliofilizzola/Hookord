package pullrequest

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/discord/colors"
)

type EventPayload struct {
	Action      string
	PullRequest *github.PullRequest
	Sender      *github.User
	Repository  *github.Repository
}

func Handle(ctx context.Context, payload *EventPayload, repo domain.MessageRepository, discord domain.DiscordProvider, channelID string) error {
	entityID := strconv.FormatInt(payload.PullRequest.GetID(), 10)

	mapping, err := repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	embed := BuildEmbed(payload)

	if mapping == nil {
		msgID, err := discord.SendMessage(ctx, channelID, embed)
		if err != nil {
			return err
		}

		return repo.SaveMapping(ctx, &domain.MessageMapping{
			DiscordMessageID: msgID,
			DiscordChannelID: channelID,
			Repository:       payload.Repository.GetFullName(),
			EntityID:         entityID,
			LastStatus:       payload.PullRequest.GetState(),
		})
	}

	return discord.EditMessage(ctx, mapping.DiscordChannelID, mapping.DiscordMessageID, embed)
}

func BuildEmbed(payload *EventPayload) *discordgo.MessageEmbed {
	pr := payload.PullRequest
	color := colors.Blue

	status := pr.GetState()
	if pr.GetDraft() {
		status = domain.PullRequestStateDraft
		color = colors.Grey
	} else if pr.GetMerged() {
		status = domain.PullRequestStateMerged
		color = colors.Purple
	} else if status == domain.PullRequestStateClosed {
		color = colors.Red
	} else if status == domain.PullRequestStateOpen {
		color = colors.Green
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Pull Request #" + strconv.Itoa(pr.GetNumber()) + " - " + pr.GetTitle(),
		URL:         pr.GetHTMLURL(),
		Description: pr.GetBody(),
		Color:       color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    payload.Sender.GetLogin(),
			IconURL: payload.Sender.GetAvatarURL(),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  status,
				Inline: true,
			},
			{
				Name:   "Repository",
				Value:  payload.Repository.GetFullName(),
				Inline: true,
			},
			{
				Name:   "Branch",
				Value:  pr.GetBase().GetRef() + " <- " + pr.GetHead().GetRef(),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "GitHub ↔ Discord Notification Hookord",
		},
	}

	return embed
}
