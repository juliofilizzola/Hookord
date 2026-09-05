package discord

import (
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func BuildIssueColor(issue *github.Issue) int {
	status := issue.GetState()
	switch status {
	case domain.IssueStateClosed:
		return ColorGrey
	case domain.IssueStateOpen:
		return ColorGreen
	default:
		return ColorOrange
	}
}

func BuildPullRequestDetectType(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	if strings.Contains(title, FIX) {
		return domain.TypeFix
	}
	if strings.Contains(title, HOT) {
		return domain.TypeHot
	}
	if strings.Contains(title, DOC) {
		return domain.TypeDoc
	}
	if strings.Contains(title, CHORE) {
		return domain.TypeChore
	}
	return domain.TypeOther
}

func BuildIssueContent(issue *github.Issue) string {
	content := ""
	if BuildPullRequestDetectType(issue.GetTitle()) == domain.TypeHot && issue.GetState() == domain.IssueStateOpen {
		content = "@everyone"
	}
	return content
}

func BuildIssueNameRepository(payload *integrations.IssueEvent) string {
	return payload.Repository.GetFullName()
}

func BuildIssueAvatarURL(payload *integrations.IssueEvent) string {
	return payload.Repository.GetOwner().GetAvatarURL()
}

func BuildIssueTitle(issue *github.Issue) string {
	return "Issue #" + strconv.Itoa(issue.GetNumber()) + " + " + issue.GetTitle()
}

func BuildIssueName(payload *integrations.IssueEvent) string {
	return payload.Sender.GetName()
}

func BuildIssueAuthorIconURL(payload *integrations.IssueEvent) string {
	return payload.Sender.GetAvatarURL()
}
