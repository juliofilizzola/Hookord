package slack

import (
	"github.com/juliofiliizzola/hookord/internal/integrations"
	"github.com/slack-go/slack"
)

func BuildPullRequestAttachment(payload *integrations.PullRequestEvent) slack.Attachment {
	pr := payload.PullRequest

	authorContext := BuildAuthorContextPullRequest(payload)

	prTitleLink := BuildPullRequestContent(pr)

	prThumbnailAccessory := BuildThumbnailAccessoryPullRequest(payload)

	titleSection := slack.NewSectionBlock(prTitleLink, nil, prThumbnailAccessory)

	statusField := BuildStatusPullRequest(pr)
	repositoryField := BuildRepositoryPullRequest(pr)

	statsField := BuildStatsPullRequest(pr)

	reviewersField := BuildReviewsPullRequest(pr)

	assigneesField := BuildAssigneesPullRequest(pr)

	labelsField := BuildLabesPullRequest(pr)
	// todo: fazer posteriormente
	totalReviewsField := slack.NewTextBlockObject("mrkdwn", "*Total de reviews*\n0", false, false)
	totalUsersField := slack.NewTextBlockObject("mrkdwn", "*Total de usuários que fizeram review*\n0", false, false)

	detailsFields := []*slack.TextBlockObject{
		statusField,
		repositoryField,
		statsField,
		reviewersField,
		assigneesField,
		labelsField,
		totalReviewsField,
		totalUsersField,
	}
	detailsSection := slack.NewSectionBlock(nil, detailsFields, nil)

	branchSection := BuildBranchPullRequest(pr)
	footerContext := BuildFooterPullRequest()

	color := BuildPullRequestColor(payload)

	data := slack.Attachment{
		Color: color,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				authorContext,
				titleSection,
				detailsSection,
				branchSection,
				footerContext,
			},
		},
	}

	return data
}
