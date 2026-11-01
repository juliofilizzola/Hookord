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
type Closer interface {
	Close() error
}

func NewPullRequestEntry(integration, channelID string, event *PullRequestEvent) Entry {
	return Entry{
		Integration:    integration,
		EntityID:       event.PullRequest.GetID(),
		ChannelID:      channelID,
		Repository:     event.Repository.GetFullName(),
		Status:         event.PullRequest.GetState(),
		TotalReviews:   event.TotalReviews,
		TotalReviewers: event.TotalReviewers,
	}
}

func NewIssueEntry(integration, channelID string, event *IssueEvent) Entry {
	return Entry{
		Integration: integration,
		EntityID:    event.Issue.GetID(),
		ChannelID:   channelID,
		Repository:  event.Repository.GetFullName(),
		Status:      event.Issue.GetState(),
	}
}
