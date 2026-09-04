package discord

import (
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func BuildPullRequestContent(pr *github.PullRequest) string {
	content := "Faça o CodeReview"
	if BuildPullRequestType(pr.GetTitle()) == domain.TypeHot && pr.GetState() == domain.PullRequestStateOpen {
		content = "@everyone"
	}
	return content
}

func BuildPullRequestNameRepository(pr *integrations.PullRequestEvent) string {
	return pr.Repository.GetFullName()
}

func BuildPullRequestName(pr *integrations.PullRequestEvent) string {
	return pr.Sender.GetName()
}

func BuildPullRequestAvatarURL(pr *integrations.PullRequestEvent) string {
	return pr.Repository.GetOwner().GetAvatarURL()
}

func BuildPullRequestIconURL(pr *integrations.PullRequestEvent) string {
	return pr.Sender.GetAvatarURL()
}

func BuildPullRequestColor(payload *integrations.PullRequestEvent) int {
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

	switch BuildPullRequestType(pr.GetTitle()) {
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

func BuildPullRequestStatus(pr *github.PullRequest) string {
	status := pr.GetState()

	if pr.GetDraft() {
		status = domain.PullRequestStateDraft
	} else if pr.GetMerged() {
		status = domain.PullRequestStateMerged
	}

	return status
}

func BuildPullRequestType(title string) string {
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

func BuildPullRequestAssignees(pr *github.PullRequest) string {
	var assignees []string
	for _, a := range pr.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	return strings.Join(assignees, ", ")
}

func BuildPullRequestLabels(pr *github.PullRequest) string {
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, l.GetName())
	}
	return strings.Join(labels, ", ")
}

func BuildPullRequestTitle(pr *github.PullRequest) string {
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

func BuildPullRequestStats(pr *github.PullRequest) string {
	return "++" + strconv.Itoa(pr.GetAdditions()) + " --" + strconv.Itoa(pr.GetDeletions())
}

func BuildPullRequestBranch(pr *github.PullRequest) string {
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

func BuildPullRequestReviewers(pr *github.PullRequest) string {
	var reviewers []string
	for _, r := range pr.RequestedReviewers {
		reviewers = append(reviewers, r.GetLogin())
	}
	return strings.Join(reviewers, ", ")
}
