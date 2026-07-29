package issues

import (
	"context"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/discord/colors"
)

type EventPayload struct {
	Action     string
	Issue      *github.Issue
	Sender     *github.User
	Repository *github.Repository
}

func Handle(ctx context.Context, payload *EventPayload, repo domain.MessageRepository, discord domain.DiscordProvider, channelID string) error {
	entityID := strconv.FormatInt(payload.Issue.GetID(), 10)

	mapping, err := repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	embed := BuildEmbed(payload)

	content := ""
	if DetectType(payload.Issue.GetTitle()) == domain.TypeHot && payload.Issue.GetState() == domain.IssueStateOpen {
		content = "@everyone"
	}

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
			LastStatus:       payload.Issue.GetState(),
		})
	}

	return discord.EditMessage(ctx, mapping.DiscordChannelID, mapping.DiscordMessageID, embed)
}

func BuildEmbed(payload *EventPayload) *discordgo.MessageEmbed {
	issue := payload.Issue
	color := colors.Orange

	status := issue.GetState()
	if status == domain.IssueStateClosed {
		color = colors.Grey
	} else if status == domain.IssueStateOpen {
		color = colors.Green
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Issue #" + strconv.Itoa(issue.GetNumber()) + " - " + issue.GetTitle(),
		URL:         issue.GetHTMLURL(),
		Description: issue.GetBody(),
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
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: payload.Repository.GetOwner().GetAvatarURL(),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "GitHub ↔ Discord Notification Hookord",
			IconURL: "https://raw.githubusercontent.com/juliofiliizzola/hookord/main/asserts/hookord_github.2.png",
		},
	}

	return embed
}

func DetectType(title string) string {
	title = strings.ToLower(title)
	if strings.Contains(title, "fix") {
		return domain.TypeFix
	}
	if strings.Contains(title, "hot") {
		return domain.TypeHot
	}
	if strings.Contains(title, "doc") {
		return domain.TypeDoc
	}
	if strings.Contains(title, "chore") {
		return domain.TypeChore
	}
	return domain.TypeOther
}
