package integrations

import (
	"context"

	"github.com/google/go-github/v60/github"
)

type PullRequestEvent struct {
	Action         string
	PullRequest    *github.PullRequest
	Sender         *github.User
	Repository     *github.Repository
	TotalReviews   int
	TotalReviewers int
}

type IssueEvent struct {
	Action     string
	Issue      *github.Issue
	Sender     *github.User
	Repository *github.Repository
}

type Integration interface {
	Name() string
	HandlePullRequest(ctx context.Context, event *PullRequestEvent) error
	HandleIssue(ctx context.Context, event *IssueEvent) error
}
