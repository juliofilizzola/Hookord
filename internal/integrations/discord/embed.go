package discord

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

const (
	footerText    = "GitHub ↔ Discord Notification Hookord"
	footerIconURL = "https://hookord-bp.s3.us-east-1.amazonaws.com/hookord_github.2.png"
)

// BuildPullRequestContent returns the text message to be sent alongside the PR embed (e.g., @everyone for hot open PRs).
func BuildPullRequestContent(pr *github.PullRequest) string {
	content := "Faça o CodeReview"
	if buildType(pr.GetTitle()) == domain.TypeHot && pr.GetState() == domain.PullRequestStateOpen {
		content = "@everyone"
	}
	return content
}

// BuildPullRequestEmbed creates a Discord message embed representing a GitHub Pull Request event.
func BuildPullRequestEmbed(payload *integrations.PullRequestEvent) *discordgo.MessageEmbed {
	pr := payload.PullRequest
	status := buildStatus(pr)
	color := buildColor(payload)

	var authorName, authorIconURL string
	if payload.Sender != nil {
		authorName = payload.Sender.GetLogin()
		authorIconURL = payload.Sender.GetAvatarURL()
	}

	var repoFullName, repoOwnerAvatarURL string
	if payload.Repository != nil {
		repoFullName = payload.Repository.GetFullName()
		if payload.Repository.GetOwner() != nil {
			repoOwnerAvatarURL = payload.Repository.GetOwner().GetAvatarURL()
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       buildTitle(pr),
		URL:         pr.GetHTMLURL(),
		Description: pr.GetBody(),
		Color:       color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    authorName,
			IconURL: authorIconURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  status,
				Inline: true,
			},
			{
				Name:   "Repository",
				Value:  repoFullName,
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
			URL: repoOwnerAvatarURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    footerText,
			IconURL: footerIconURL,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return embed
}

// BuildIssueContent returns the text message for an Issue event (e.g., @everyone for hot open issues).
func BuildIssueContent(issue *github.Issue) string {
	content := ""
	if detectIssueType(issue.GetTitle()) == domain.TypeHot && issue.GetState() == domain.IssueStateOpen {
		content = "@everyone"
	}
	return content
}

// BuildIssueEmbed creates a Discord message embed representing a GitHub Issue event.
func BuildIssueEmbed(payload *integrations.IssueEvent) *discordgo.MessageEmbed {
	issue := payload.Issue
	color := ColorOrange

	status := issue.GetState()
	if status == domain.IssueStateClosed {
		color = ColorGrey
	} else if status == domain.IssueStateOpen {
		color = ColorGreen
	}

	var authorName, authorIconURL string
	if payload.Sender != nil {
		authorName = payload.Sender.GetLogin()
		authorIconURL = payload.Sender.GetAvatarURL()
	}

	var repoFullName, repoOwnerAvatarURL string
	if payload.Repository != nil {
		repoFullName = payload.Repository.GetFullName()
		if payload.Repository.GetOwner() != nil {
			repoOwnerAvatarURL = payload.Repository.GetOwner().GetAvatarURL()
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Issue #" + strconv.Itoa(issue.GetNumber()) + " - " + issue.GetTitle(),
		URL:         issue.GetHTMLURL(),
		Description: issue.GetBody(),
		Color:       color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    authorName,
			IconURL: authorIconURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  status,
				Inline: true,
			},
			{
				Name:   "Repository",
				Value:  repoFullName,
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: repoOwnerAvatarURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    footerText,
			IconURL: footerIconURL,
		},
	}

	return embed
}

func buildColor(payload *integrations.PullRequestEvent) int {
	pr := payload.PullRequest

	if pr.GetDraft() {
		return ColorGrey
	}

	if pr.GetMerged() {
		return ColorPurple
	}

	if pr.GetState() == domain.PullRequestStateClosed {
		return ColorDarkGrey
	}

	switch buildType(pr.GetTitle()) {
	case domain.TypeFix:
		return ColorOrange
	case domain.TypeHot:
		return ColorRed
	case domain.TypeDoc:
		return ColorBlue
	case domain.TypeChore:
		return ColorYellow
	default:
		return ColorGreen
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

func detectIssueType(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
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
	baseRef := ""
	if pr.GetBase() != nil {
		baseRef = pr.GetBase().GetRef()
	}
	headRef := ""
	if pr.GetHead() != nil {
		headRef = pr.GetHead().GetRef()
	}
	return "Pull Request #" + strconv.Itoa(pr.GetNumber()) + " - " + pr.GetTitle() + "[" + baseRef + " <- " + headRef + "]"
}

func buildStats(pr *github.PullRequest) string {
	return "++" + strconv.Itoa(pr.GetAdditions()) + " --" + strconv.Itoa(pr.GetDeletions())
}

func buildBranch(pr *github.PullRequest) string {
	baseRef := ""
	if pr.GetBase() != nil {
		baseRef = pr.GetBase().GetRef()
	}
	headRef := ""
	if pr.GetHead() != nil {
		headRef = pr.GetHead().GetRef()
	}
	return baseRef + " <- " + headRef
}
