package slack

import (
	"github.com/juliofiliizzola/hookord/internal/integrations"
	"github.com/slack-go/slack"
)

func BuildPullRequestAttachment(payload *integrations.PullRequestEvent) slack.Attachment {
	pr := payload.PullRequest

	//authorAvatar := slack.NewImageBlockElement("https://avatars.githubusercontent.com/u/sua-imagem-aqui", "Author Avatar")
	authorAvatar := BuildPullRequestThumbnail(payload)
	authorName := slack.NewTextBlockObject("mrkdwn", "**juliofilizzola**", false, false)
	authorContext := slack.NewContextBlock("author_context", authorAvatar, authorName)

	prTitleLink := BuildPullRequestContent(pr)
	////prThumbnail := slack.NewImageBlockElement("https://avatars.githubusercontent.com/u/imagem-lateral-aqui", "PR Thumbnail")
	//prThumbnail := BuildPullRequestThumbnail(payload)
	//prThumbnailAccessory := slack.NewAccessory(prThumbnail)

	prThumbnail := BuildPullRequestThumbnail(payload)
	prThumbnailAccessory := slack.NewAccessory(prThumbnail)
	titleSection := slack.NewSectionBlock(prTitleLink, nil, prThumbnailAccessory)

	statusField := slack.NewTextBlockObject("mrkdwn", "*Status*\nopen", false, false)
	repositoryField := slack.NewTextBlockObject("mrkdwn", "*Repository*\njuliofilizzola/Hookord", false, false)
	statsField := slack.NewTextBlockObject("mrkdwn", "*Stats*\n++8 --0", false, false)
	reviewersField := slack.NewTextBlockObject("mrkdwn", "*Reviewers*\n-", false, false)
	assigneesField := slack.NewTextBlockObject("mrkdwn", "*Assignees*\njuliofilizzola", false, false)
	labelsField := slack.NewTextBlockObject("mrkdwn", "*Labels*\nenhancement", false, false)
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

	branchField := slack.NewTextBlockObject("mrkdwn", "*Branch*\nmain <- 40-feat-adicionar-provider-do-slack", false, false)
	branchSection := slack.NewSectionBlock(branchField, nil, nil)

	footerIcon := slack.NewImageBlockElement("https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png", "GitHub Icon")
	footerText := slack.NewTextBlockObject("mrkdwn", "GitHub ↔ Discord Notification Hookord • Hoje às 19:58", false, false)
	footerContext := slack.NewContextBlock("footer_context", footerIcon, footerText)

	data := slack.Attachment{
		Color: "#2eb886",
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
