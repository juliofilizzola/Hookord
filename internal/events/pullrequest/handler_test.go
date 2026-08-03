package pullrequest

import (
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/discord/colors"
)

func TestDetectType(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{name: "fix", title: "fix: correct panic", expected: domain.TypeFix},
		{name: "hot uppercase", title: "HOT urgent patch", expected: domain.TypeHot},
		{name: "doc", title: "docs: update readme", expected: domain.TypeDoc},
		{name: "chore", title: "chore: bump dependency", expected: domain.TypeChore},
		{name: "other", title: "feature: add dashboard", expected: domain.TypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectType(tt.title); got != tt.expected {
				t.Fatalf("DetectType(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestDetectStaus(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		draft    bool
		merged   bool
		expected int
	}{
		{name: "draft", state: domain.PullRequestStateOpen, draft: true, expected: colors.Grey},
		{name: "merged", state: domain.PullRequestStateClosed, merged: true, expected: colors.Purple},
		{name: "closed", state: domain.PullRequestStateClosed, expected: colors.Red},
		{name: "open", state: domain.PullRequestStateOpen, expected: colors.Green},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &EventPayload{
				PullRequest: &github.PullRequest{
					State:  github.String(tt.state),
					Draft:  github.Bool(tt.draft),
					Merged: github.Bool(tt.merged),
				},
			}

			if got := DetectStaus(payload); got != tt.expected {
				t.Fatalf("DetectStaus() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestBuildEmbed_UsesDraftStatusAndHotColor(t *testing.T) {
	payload := &EventPayload{
		PullRequest: &github.PullRequest{
			Number:    github.Int(17),
			Title:     github.String("hotfix: urgent production fix"),
			Body:      github.String("fix issue"),
			HTMLURL:   github.String("https://example.com/pr/17"),
			State:     github.String(domain.PullRequestStateOpen),
			Draft:     github.Bool(true),
			Additions: github.Int(10),
			Deletions: github.Int(2),
			Base:      &github.PullRequestBranch{Ref: github.String("main")},
			Head:      &github.PullRequestBranch{Ref: github.String("fix/prod")},
		},
		Sender: &github.User{
			Login:     github.String("julio"),
			AvatarURL: github.String("https://example.com/avatar.png"),
		},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner: &github.User{
				AvatarURL: github.String("https://example.com/repo-avatar.png"),
			},
		},
	}

	embed := BuildEmbed(payload)

	if embed.Color != colors.Red {
		t.Fatalf("embed.Color = %d, want %d", embed.Color, colors.Red)
	}
	if got := embed.Fields[0].Value; got != domain.PullRequestStateDraft {
		t.Fatalf("status field = %q, want %q", got, domain.PullRequestStateDraft)
	}
}
