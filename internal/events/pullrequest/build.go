package pullrequest

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/discord/colors"
)

func buildColor(payload *EventPayload) int {
	pr := payload.PullRequest

	if pr.GetDraft() {
		return colors.Grey
	}

	if pr.GetMerged() {
		return colors.Purple
	}

	if pr.GetState() == domain.PullRequestStateClosed {
		return colors.DarkGrey
	}

	switch buildType(pr.GetTitle()) {
	case domain.TypeFix:
		return colors.Orange
	case domain.TypeHot:
		return colors.Red
	case domain.TypeDoc:
		return colors.Blue
	case domain.TypeChore:
		return colors.Yellow
	default:
		return colors.Green
	}
}

func buildType(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))

	switch {
	case strings.HasPrefix(title, "feat"):
		return domain.TypeFeat
	case strings.HasPrefix(title, "fix"):
		return domain.TypeFix
	case strings.HasPrefix(title, "hot"):
		return domain.TypeHot
	case strings.HasPrefix(title, "doc"):
		return domain.TypeDoc
	case strings.HasPrefix(title, "chore"):
		return domain.TypeChore
	default:
		return domain.TypeOther
	}
}

func BuildContent(pr *github.PullRequest) string {
	content := "Faça o CodeReview"
	if buildType(pr.GetTitle()) == domain.TypeHot && pr.GetState() == domain.PullRequestStateOpen {
		content = "@everyone"
	}
	return content
}

func buildStatus(pr *github.PullRequest) string {
	status := pr.GetState()

	if pr.GetDraft() {
		status = domain.PullRequestStateDraft
	} else if pr.GetMerged() {
		status = domain.PullRequestStateMerged
	}

	return status
}

func buildReviewers(pr *github.PullRequest) string {
	var reviewers []string
	for _, r := range pr.RequestedReviewers {
		reviewers = append(reviewers, r.GetLogin())
	}
	return strings.Join(reviewers, ", ")
}

func buildAssignees(pr *github.PullRequest) string {
	var assignees []string
	for _, a := range pr.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	return strings.Join(assignees, ", ")
}

func buildLabels(pr *github.PullRequest) string {
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, l.GetName())
	}
	return strings.Join(labels, ", ")
}

func buildTitle(pr *github.PullRequest) string {
	return "Pull Request #" + strconv.Itoa(pr.GetNumber()) + " - " + pr.GetTitle() + "[" + pr.GetBase().GetRef() + " <- " + pr.GetHead().GetRef() + "]"
}

func buildStats(pr *github.PullRequest) string {
	return "++" + strconv.Itoa(pr.GetAdditions()) + " --" + strconv.Itoa(pr.GetDeletions())
}

func buildBranch(pr *github.PullRequest) string {
	return pr.GetBase().GetRef() + " <- " + pr.GetHead().GetRef()
}

func BuildEmbed(payload *EventPayload) *discordgo.MessageEmbed {
	pr := payload.PullRequest

	status := buildStatus(pr)

	color := buildColor(payload)

	embed := &discordgo.MessageEmbed{
		Title:       buildTitle(pr),
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
				Value:  buildStats(pr),
				Inline: true,
			},
			{
				Name:   "Reviewers",
				Value:  buildReviewers(pr),
				Inline: true,
			},
			{
				Name:   "Assignees",
				Value:  buildAssignees(pr),
				Inline: true,
			},
			{
				Name:   "Labels",
				Value:  buildLabels(pr),
				Inline: true,
			},
			{
				Name:   "Total de reviews",
				Value:  strconv.Itoa(payload.TotalReviews),
				Inline: true,
			},
			{
				Name:   "Total de usuários que fizeram review",
				Value:  strconv.Itoa(payload.TotalReviewers),
				Inline: true,
			},
			{
				Name:   "Branch",
				Value:  buildBranch(pr),
				Inline: false,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: payload.Repository.GetOwner().GetAvatarURL(),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "GitHub ↔ Discord Notification Hookord",
			IconURL: "https://hookord-bp.s3.us-east-1.amazonaws.com/hookord_github.2.png",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return embed
}
