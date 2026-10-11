package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func BuildIssueMessage(payload *integrations.IssueEvent) string {
	issue := payload.Issue
	var sb strings.Builder

	// Header
	authorName := "Unknown"
	if issue.User != nil {
		authorName = issue.User.GetName()
		if authorName == "" {
			authorName = issue.User.GetLogin()
		}
	}
	sb.WriteString(fmt.Sprintf("🔨 <b>%s</b>\n", escapeHTML(authorName)))

	// Title
	issueTitleUnformatted := "Issue #" + strconv.Itoa(issue.GetNumber()) + " - " + issue.GetTitle()
	issueTitleEscaped := escapeHTML(issueTitleUnformatted)
	sb.WriteString(fmt.Sprintf("<a href=\"%s\"><b>%s</b></a>\n\n", issue.GetHTMLURL(), issueTitleEscaped))

	// Details
	sb.WriteString(fmt.Sprintf("<b>Status:</b> %s\n", issue.GetState()))

	repoName := ""
	if payload.Repository != nil {
		repoName = payload.Repository.GetFullName()
	}
	sb.WriteString(fmt.Sprintf("<b>Repository:</b> %s\n", escapeHTML(repoName)))

	// Assignees
	var assignees []string
	for _, a := range issue.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	if len(assignees) > 0 {
		sb.WriteString(fmt.Sprintf("<b>Assignees:</b> %s\n", escapeHTML(strings.Join(assignees, ", "))))
	} else {
		sb.WriteString("<b>Assignees:</b> None\n")
	}

	// Labels
	var labels []string
	for _, a := range issue.Labels {
		labels = append(labels, a.GetName())
	}
	if len(labels) > 0 {
		sb.WriteString(fmt.Sprintf("<b>Labels:</b> %s\n", escapeHTML(strings.Join(labels, ", "))))
	} else {
		sb.WriteString("<b>Labels:</b> None\n")
	}

	return sb.String()
}
