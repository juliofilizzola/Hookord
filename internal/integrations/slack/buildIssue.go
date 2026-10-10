package slack

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/common/utils"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
	"github.com/slack-go/slack"
)

func BuildIssueAttachment(payload *integrations.IssueEvent) slack.Attachment {
	issue := payload.Issue

	authorContext := BuildAuthorContextIssue(payload)
	titleSection := slack.NewSectionBlock(BuildIssueContent(issue), nil, BuildThumbnailAccessoryIssue(payload))

	statusField := BuildStatusIssue(issue)
	repositoryField := BuildRepositoryIssue(payload.Repository)
	assigneesField := BuildAssigneesIssue(issue)
	labelsField := BuildLabelsIssue(issue)

	detailsFields := []*slack.TextBlockObject{
		statusField,
		repositoryField,
		assigneesField,
		labelsField,
	}
	detailsSection := slack.NewSectionBlock(nil, detailsFields, nil)

	footerContext := BuildFooterPullRequest() // Reusing the same footer logic

	color := BuildIssueColor(payload)

	return slack.Attachment{
		Color: color,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				authorContext,
				titleSection,
				detailsSection,
				footerContext,
			},
		},
	}
}

func BuildIssueContent(issue *github.Issue) *slack.TextBlockObject {
	issueTitleUnformatted := "Issue #" + strconv.Itoa(issue.GetNumber()) + " - " + issue.GetTitle()
	issueTitleEscaped := escapeSlackMarkdownReservedCharacters(issueTitleUnformatted)
	slackFormattedLink := "<" + issue.GetHTMLURL() + "|*" + issueTitleEscaped + "*>"
	return slack.NewTextBlockObject(ElementType, slackFormattedLink, false, false)
}

func BuildThumbnailAccessoryIssue(payload *integrations.IssueEvent) *slack.Accessory {
	thumb := slack.NewImageBlockElement(payload.Repository.GetOwner().GetAvatarURL(), payload.Repository.GetOwner().GetName())
	return slack.NewAccessory(thumb)
}

func BuildAuthorContextIssue(payload *integrations.IssueEvent) *slack.ContextBlock {
	authorAvatar := slack.NewImageBlockElement(payload.Issue.User.GetAvatarURL(), payload.Issue.User.GetName())
	authorText := slack.NewTextBlockObject(ElementType, fmt.Sprintf("*%s*", payload.Issue.User.GetName()), false, false)
	return slack.NewContextBlock(AuthorContext, authorAvatar, authorText)
}

func BuildStatusIssue(issue *github.Issue) *slack.TextBlockObject {
	status := fmt.Sprintf("*Status*\n%s", issue.GetState())
	return slack.NewTextBlockObject(ElementType, status, false, false)
}

func BuildAssigneesIssue(issue *github.Issue) *slack.TextBlockObject {
	var assignees []string
	for _, a := range issue.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	if len(assignees) == 0 {
		return slack.NewTextBlockObject(ElementType, "*Assignees*:\nNone", false, false)
	}
	assigneesFormat := fmt.Sprintf("*Assignees*:\n%s", strings.Join(assignees, ", "))
	return slack.NewTextBlockObject(ElementType, assigneesFormat, false, false)
}

func BuildLabelsIssue(issue *github.Issue) *slack.TextBlockObject {
	var labels []string
	for _, a := range issue.Labels {
		labels = append(labels, a.GetName())
	}
	if len(labels) == 0 {
		return slack.NewTextBlockObject(ElementType, "*Labels*:\nNone", false, false)
	}
	labelsFormat := fmt.Sprintf("*Labels*:\n%s", strings.Join(labels, ", "))
	return slack.NewTextBlockObject(ElementType, labelsFormat, false, false)
}

func BuildRepositoryIssue(repo *github.Repository) *slack.TextBlockObject {
	repository := fmt.Sprintf("*Repository*:\n%s", repo.GetFullName())
	return slack.NewTextBlockObject(ElementType, repository, false, false)
}

func BuildIssueColor(payload *integrations.IssueEvent) string {
	issue := payload.Issue

	if issue.GetState() == domain.PullRequestStateClosed {
		return utils.ColorDarkGreyText
	}

	switch utils.TypePullRequest(issue.GetTitle()) { // Reusing TypePullRequest since the naming convention is likely the same
	case domain.TypeFix:
		return utils.ColorOrangeText
	case domain.TypeHot:
		return utils.ColorRedText
	case domain.TypeDoc:
		return utils.ColorBlueText
	case domain.TypeChore:
		return utils.ColorYellowText
	default:
		return utils.ColorGreenText
	}
}
