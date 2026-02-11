package providers

import (
	"github.com/juliofilizzola/Hookord/internal/core"
	"github.com/juliofilizzola/Hookord/internal/providers/github"
)

type InputProvider interface {
	Parse(eventType string, payload []byte) (core.Event, error)

	ValidateSignature(payload []byte, signature string, secret string) error
}

type OutputProvider interface {
	SendMessage(event core.Event) error

	Name() string
}

type GithubEventType string

const (
	GithubPushEvent        GithubEventType = "push"
	GithubIssueEvent       GithubEventType = "issues"
	GithubPullRequestEvent GithubEventType = "pull_request"
	GithubReleaseEvent     GithubEventType = "release"
	GithubWorkflowEvent    GithubEventType = "workflow_dispatch"
	GithubPingEvent        GithubEventType = "ping"
	GithubDeploymentEvent  GithubEventType = "deployment"
)
