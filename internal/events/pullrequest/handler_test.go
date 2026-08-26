package pullrequest

import (
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/discord/colors"
)

func TestDetectStatus(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		draft    bool
		merged   bool
		expected int
	}{
		{name: "draft", state: domain.PullRequestStateOpen, draft: true, expected: colors.Grey},
		{name: "merged", state: domain.PullRequestStateClosed, merged: true, expected: colors.Purple},
		{name: "closed", state: domain.PullRequestStateClosed, expected: colors.Purple},
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

			if got := buildColor(payload); got != tt.expected {
				t.Fatalf("DetectStaus() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestBuildEmbed_UsesDraftStatusAndHotColor(t *testing.T) {
	payload := &EventPayload{
		PullRequest: &github.PullRequest{
			Number:    github.Int(17),
			Title:     github.String("HOT: urgent production patch"),
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

func TestBuildEmbed_IncludesReviewCounts(t *testing.T) {
	payload := &EventPayload{
		PullRequest: &github.PullRequest{
			Number:    github.Int(31),
			Title:     github.String("Refactor deployment"),
			State:     github.String(domain.PullRequestStateOpen),
			Additions: github.Int(100),
			Deletions: github.Int(20),
			Base:      &github.PullRequestBranch{Ref: github.String("main")},
			Head:      &github.PullRequestBranch{Ref: github.String("feature")},
		},
		Sender: &github.User{
			Login:     github.String("julio"),
			AvatarURL: github.String("https://example.com/avatar.png"),
		},
		Repository: &github.Repository{
			FullName: github.String("juliofilizzola/Hookord"),
			Owner: &github.User{
				AvatarURL: github.String("https://example.com/repo-avatar.png"),
			},
		},
		TotalReviews:   5,
		TotalReviewers: 3,
	}

	embed := BuildEmbed(payload)

	var foundTotalReviews, foundTotalReviewers bool
	for _, field := range embed.Fields {
		if field.Name == "Total de reviews" {
			foundTotalReviews = true
			if field.Value != "5" {
				t.Errorf("Total de reviews field value = %q, want %q", field.Value, "5")
			}
		}
		if field.Name == "Total de usuários que fizeram review" {
			foundTotalReviewers = true
			if field.Value != "3" {
				t.Errorf("Total de usuários que fizeram review field value = %q, want %q", field.Value, "3")
			}
		}
	}

	if !foundTotalReviews {
		t.Errorf("field 'Total de reviews' missing from embed")
	}
	if !foundTotalReviewers {
		t.Errorf("field 'Total de usuários que fizeram review' missing from embed")
	}
}
