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

func escapeSlackMarkdownReservedCharacters(text string) string {
	textEscapingAmpersand := strings.ReplaceAll(text, "&", "&amp;")
	textEscapingLessThan := strings.ReplaceAll(textEscapingAmpersand, "<", "&lt;")
	textEscapingGreaterThan := strings.ReplaceAll(textEscapingLessThan, ">", "&gt;")
	return textEscapingGreaterThan
}

func BuildPullRequestContent(pullRequest *github.PullRequest) *slack.TextBlockObject {
	baseBranchRef := ""
	if pullRequest.GetBase() != nil {
		baseBranchRef = pullRequest.GetBase().GetRef()
	}

	headBranchRef := ""
	if pullRequest.GetHead() != nil {
		headBranchRef = pullRequest.GetHead().GetRef()
	}

	pullRequestTitleUnformatted := "Pull Request #" + strconv.Itoa(pullRequest.GetNumber()) + " - " + pullRequest.GetTitle() + " [" + baseBranchRef + " <- " + headBranchRef + "]"
	pullRequestTitleEscaped := escapeSlackMarkdownReservedCharacters(pullRequestTitleUnformatted)

	pullRequestURL := pullRequest.GetHTMLURL()

	slackFormattedLink := "<" + pullRequestURL + "|*" + pullRequestTitleEscaped + "*>"

	return slack.NewTextBlockObject(ElementType, slackFormattedLink, false, false)
}

func BuildPullRequestThumbnail(pr *integrations.PullRequestEvent) *slack.ImageBlockElement {
	return slack.NewImageBlockElement(pr.Repository.GetOwner().GetAvatarURL(), pr.Repository.GetOwner().GetName())
}

func BuildThumbnailAccessoryPullRequest(pr *integrations.PullRequestEvent) *slack.Accessory {
	thumb := BuildPullRequestThumbnail(pr)
	return slack.NewAccessory(thumb)
}

func BuildAuthorContextPullRequest(pr *integrations.PullRequestEvent) *slack.ContextBlock {
	authorAvatar := slack.NewImageBlockElement(pr.PullRequest.User.GetAvatarURL(), pr.PullRequest.User.GetName())
	authorText := slack.NewTextBlockObject(ElementType, fmt.Sprintf("*%s*", pr.PullRequest.User.GetName()), false, false)

	return slack.NewContextBlock(AuthorContext, authorAvatar, authorText)
}

func BuildFooterPullRequest() *slack.ContextBlock {
	footerIcon := slack.NewImageBlockElement(utils.FooterIconURL, utils.FooterIconAlt)
	footerContext := slack.NewTextBlockObject(ElementType, utils.FooterText, false, false)
	return slack.NewContextBlock(FooterContext, footerIcon, footerContext)
}

func BuildBranchPullRequest(pr *github.PullRequest) *slack.SectionBlock {
	baseRef := ""
	if pr.GetBase() != nil {
		baseRef = pr.GetBase().GetRef()
	}
	headRef := ""
	if pr.GetHead() != nil {
		headRef = pr.GetHead().GetRef()
	}

	branch := fmt.Sprintf("*Branch*\n%s <- %s", baseRef, headRef)
	br := slack.NewTextBlockObject(ElementType, branch, false, false)
	return slack.NewSectionBlock(br, nil, nil)
}

func BuildStatsPullRequest(pr *github.PullRequest) *slack.TextBlockObject {
	stats := fmt.Sprintf("*Stats*:\n%d files changed, %d additions, %d deletions", pr.GetChangedFiles(), pr.GetAdditions(), pr.GetDeletions())

	return slack.NewTextBlockObject(ElementType, stats, false, false)
}

func BuildStatusPullRequest(pr *github.PullRequest) *slack.TextBlockObject {
	statusEvent := pr.GetState()

	if pr.GetDraft() {
		statusEvent = domain.PullRequestStateDraft
	} else if pr.GetMerged() {
		statusEvent = domain.PullRequestStateMerged
	}

	status := fmt.Sprintf("*Status*\n%s", statusEvent)

	return slack.NewTextBlockObject(ElementType, status, false, false)
}

func BuildAssigneesPullRequest(pr *github.PullRequest) *slack.TextBlockObject {
	var assignees []string

	for _, a := range pr.Assignees {
		assignees = append(assignees, a.GetLogin())
	}

	assigneesFormat := fmt.Sprintf("*Assignees*:\n%s", strings.Join(assignees, ", "))

	return slack.NewTextBlockObject(ElementType, assigneesFormat, false, false)
}

func BuildLabesPullRequest(pr *github.PullRequest) *slack.TextBlockObject {
	var labels []string

	for _, a := range pr.Labels {
		labels = append(labels, a.GetName())
	}

	labelsFormat := fmt.Sprintf("*Labels*:\n%s", strings.Join(labels, ", "))

	return slack.NewTextBlockObject(ElementType, labelsFormat, false, false)
}

func BuildRepositoryPullRequest(pr *github.PullRequest) *slack.TextBlockObject {
	repository := fmt.Sprintf("*Repository*:\n%s", pr.GetBase().GetRepo().GetFullName())

	return slack.NewTextBlockObject(ElementType, repository, false, false)
}

func BuildReviewsPullRequest(pr *github.PullRequest) *slack.TextBlockObject {
	var Reviewers []string

	for _, a := range pr.RequestedReviewers {
		Reviewers = append(Reviewers, a.GetLogin())
	}

	ReviewersFormat := fmt.Sprintf("*Reviews*:\n%s", strings.Join(Reviewers, ", "))
	return slack.NewTextBlockObject(ElementType, ReviewersFormat, false, false)
}
