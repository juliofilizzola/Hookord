package slack

import (
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
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

	return slack.NewTextBlockObject("mrkdwn", slackFormattedLink, false, false)
}

func BuildPullRequestThumbnail(pr *integrations.PullRequestEvent) *slack.ImageBlockElement {
	return slack.NewImageBlockElement(pr.Repository.GetOwner().GetAvatarURL(), pr.Repository.GetOwner().GetName())
}
