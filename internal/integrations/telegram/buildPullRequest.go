package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}

func BuildPullRequestMessage(payload *integrations.PullRequestEvent) string {
	pr := payload.PullRequest
	var sb strings.Builder

	// Header
	authorName := "Unknown"
	if pr.User != nil {
		authorName = pr.User.GetName()
		if authorName == "" {
			authorName = pr.User.GetLogin()
		}
	}
	sb.WriteString(fmt.Sprintf("🔨 <b>%s</b>\n", escapeHTML(authorName)))

	// Title
	baseRef := ""
	if pr.GetBase() != nil {
		baseRef = pr.GetBase().GetRef()
	}
	headRef := ""
	if pr.GetHead() != nil {
		headRef = pr.GetHead().GetRef()
	}

	prTitleUnformatted := "Pull Request #" + strconv.Itoa(pr.GetNumber()) + " - " + pr.GetTitle() + " [" + baseRef + " <- " + headRef + "]"
	prTitleEscaped := escapeHTML(prTitleUnformatted)
	sb.WriteString(fmt.Sprintf("<a href=\"%s\"><b>%s</b></a>\n\n", pr.GetHTMLURL(), prTitleEscaped))

	// Details
	statusEvent := pr.GetState()
	if pr.GetDraft() {
		statusEvent = domain.PullRequestStateDraft
	} else if pr.GetMerged() {
		statusEvent = domain.PullRequestStateMerged
	}

	sb.WriteString(fmt.Sprintf("<b>Status:</b> %s\n", statusEvent))

	repoName := ""
	if pr.GetBase() != nil && pr.GetBase().GetRepo() != nil {
		repoName = pr.GetBase().GetRepo().GetFullName()
	}
	sb.WriteString(fmt.Sprintf("<b>Repository:</b> %s\n", escapeHTML(repoName)))

	sb.WriteString(fmt.Sprintf("<b>Stats:</b> %d files changed, %d additions, %d deletions\n", pr.GetChangedFiles(), pr.GetAdditions(), pr.GetDeletions()))

	// Reviewers
	var reviewers []string
	for _, a := range pr.RequestedReviewers {
		reviewers = append(reviewers, a.GetLogin())
	}
	if len(reviewers) > 0 {
		sb.WriteString(fmt.Sprintf("<b>Reviews:</b> %s\n", escapeHTML(strings.Join(reviewers, ", "))))
	}

	// Assignees
	var assignees []string
	for _, a := range pr.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	if len(assignees) > 0 {
		sb.WriteString(fmt.Sprintf("<b>Assignees:</b> %s\n", escapeHTML(strings.Join(assignees, ", "))))
	}

	// Labels
	var labels []string
	for _, a := range pr.Labels {
		labels = append(labels, a.GetName())
	}
	if len(labels) > 0 {
		sb.WriteString(fmt.Sprintf("<b>Labels:</b> %s\n", escapeHTML(strings.Join(labels, ", "))))
	}

	sb.WriteString(fmt.Sprintf("<b>Total de reviews:</b> %d\n", payload.TotalReviews))
	sb.WriteString(fmt.Sprintf("<b>Total de usuários que fizeram review:</b> %d\n", payload.TotalReviewers))

	return sb.String()
}
