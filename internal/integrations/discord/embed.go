package discord

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func BuildPullRequestEmbed(payload *integrations.PullRequestEvent) *discordgo.MessageEmbed {
	pr := payload.PullRequest
	status := BuildPullRequestStatus(pr)
	color := BuildPullRequestColor(payload)

	authorIconURL := BuildPullRequestIconURL(payload)
	authorName := BuildPullRequestName(payload)

	repoFullName := BuildPullRequestNameRepository(payload)
	repoOwnerAvatarURL := BuildPullRequestAvatarURL(payload)

	embed := &discordgo.MessageEmbed{
		Title:       BuildPullRequestTitle(pr),
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
				Value:  BuildPullRequestStats(pr),
				Inline: true,
			},
			{
				Name:   "Reviewers",
				Value:  BuildPullRequestReviewers(pr),
				Inline: true,
			},
			{
				Name:   "Assignees",
				Value:  BuildPullRequestAssignees(pr),
				Inline: true,
			},
			{
				Name:   "Labels",
				Value:  BuildPullRequestLabels(pr),
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
				Value:  BuildPullRequestBranch(pr),
				Inline: false,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: repoOwnerAvatarURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    FooterText,
			IconURL: FooterIconURL,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return embed
}

func BuildIssueEmbed(payload *integrations.IssueEvent) *discordgo.MessageEmbed {
	issue := payload.Issue
	color := BuildIssueColor(issue)

	status := issue.GetState()
	authorName := BuildIssueName(payload)
	authorIconURL := BuildIssueAuthorIconURL(payload)

	repoFullName := BuildIssueNameRepository(payload)
	repoOwnerAvatarURL := BuildIssueAvatarURL(payload)

	title := BuildIssueTitle(issue)

	embed := &discordgo.MessageEmbed{
		Title:       title,
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
			Text:    FooterText,
			IconURL: FooterIconURL,
		},
	}

	return embed
}
