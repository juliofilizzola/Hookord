package integrations

import (
	"context"

	"github.com/google/go-github/v60/github"
)

// PullRequestEvent contains the necessary context for Pull Request notifications.
type PullRequestEvent struct {
	Action         string
	PullRequest    *github.PullRequest
	Sender         *github.User
	Repository     *github.Repository
	TotalReviews   int
	TotalReviewers int
}

// IssueEvent contains the necessary context for Issue notifications.
type IssueEvent struct {
	Action     string
	Issue      *github.Issue
	Sender     *github.User
	Repository *github.Repository
}

// Integration defines the contract that any notification provider (Discord, Slack, etc.) must implement.
type Integration interface {
	Name() string
	HandlePullRequest(ctx context.Context, event *PullRequestEvent) error
	HandleIssue(ctx context.Context, event *IssueEvent) error
}
