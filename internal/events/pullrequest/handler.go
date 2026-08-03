package pullrequest

import (
	"context"
	"strconv"
	"strings"
	"time"

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

	reviewers := []string{}
	for _, reviewer := range payload.PullRequest.RequestedReviewers {
		reviewers = append(reviewers, "@"+reviewer.GetLogin())
	}

	content := ""
	if DetectType(payload.PullRequest.GetTitle()) == domain.TypeHot && payload.PullRequest.GetState() == domain.PullRequestStateOpen && !payload.PullRequest.GetDraft() {
		content = "@everyone"
	}

	if len(reviewers) > 0 && payload.PullRequest.GetState() == domain.PullRequestStateOpen {
		if content != "" {
			content += " "
		}
		content += strings.Join(reviewers, " ")
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
			LastStatus:       payload.PullRequest.GetState(),
		})
	}

	return discord.EditMessage(ctx, mapping.DiscordChannelID, mapping.DiscordMessageID, embed)
}

func BuildEmbed(payload *EventPayload) *discordgo.MessageEmbed {
	pr := payload.PullRequest

	status := pr.GetState()
	if pr.GetDraft() {
		status = domain.PullRequestStateDraft
	} else if pr.GetMerged() {
		status = domain.PullRequestStateMerged
	}

	color := DetectStaus(payload)

	if DetectType(pr.GetTitle()) == domain.TypeHot {
		color = colors.Red
	}

	var reviewers []string
	for _, r := range pr.RequestedReviewers {
		reviewers = append(reviewers, r.GetLogin())
	}
	reviewersStr := "None"
	if len(reviewers) > 0 {
		reviewersStr = strings.Join(reviewers, ", ")
	}

	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, l.GetName())
	}
	labelsStr := "None"
	if len(labels) > 0 {
		labelsStr = strings.Join(labels, ", ")
	}

	var assignees []string
	for _, a := range pr.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	assigneesStr := "None"
	if len(assignees) > 0 {
		assigneesStr = strings.Join(assignees, ", ")
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
				Inline: false,
			},
			{
				Name:   "Stats",
				Value:  "++" + strconv.Itoa(pr.GetAdditions()) + " --" + strconv.Itoa(pr.GetDeletions()),
				Inline: true,
			},
			{
				Name:   "Reviewers",
				Value:  reviewersStr,
				Inline: true,
			},
			{
				Name:   "Assignees",
				Value:  assigneesStr,
				Inline: true,
			},
			{
				Name:   "Labels",
				Value:  labelsStr,
				Inline: true,
			},
			{
				Name:   "Branch",
				Value:  pr.GetBase().GetRef() + " <- " + pr.GetHead().GetRef(),
				Inline: false,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: payload.Repository.GetOwner().GetAvatarURL(),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "GitHub ↔ Discord Notification Hookord",
			IconURL: "./../asserts/hookord_github.2.png",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return embed
}

func DetectStaus(payload *EventPayload) int {
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

	return color
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
