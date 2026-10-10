package slack

import (
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/common/utils"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

func TestBuildIssueAttachment(t *testing.T) {
	issueEvent := &integrations.IssueEvent{
		Action: "opened",
		Issue: &github.Issue{
			ID:     github.Int64(123),
			Number: github.Int(42),
			Title:  github.String("Test Issue"),
			State:  github.String(domain.PullRequestStateOpen),
			User: &github.User{
				Login:     github.String("testuser"),
				AvatarURL: github.String("https://example.com/avatar.png"),
				Name:      github.String("Test User"),
			},
			HTMLURL: github.String("https://github.com/org/repo/issues/42"),
			Assignees: []*github.User{
				{Login: github.String("user1")},
				{Login: github.String("user2")},
			},
			Labels: []*github.Label{
				{Name: github.String("bug")},
			},
		},
		Sender: &github.User{
			Login:     github.String("testuser"),
			AvatarURL: github.String("https://example.com/avatar.png"),
		},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner: &github.User{
				Name:      github.String("org"),
				AvatarURL: github.String("https://example.com/org.png"),
			},
		},
	}

	attachment := BuildIssueAttachment(issueEvent)

	if len(attachment.Blocks.BlockSet) != 4 {
		t.Errorf("Expected 4 blocks, got %d", len(attachment.Blocks.BlockSet))
	}
}

func TestBuildIssueColor(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		title    string
		expected string
	}{
		{
			name:     "Closed issue",
			state:    domain.PullRequestStateClosed,
			title:    "Some issue",
			expected: utils.ColorDarkGreyText,
		},
		{
			name:     "Fix issue",
			state:    domain.PullRequestStateOpen,
			title:    "fix: some bug",
			expected: utils.ColorOrangeText,
		},
		{
			name:     "Default issue",
			state:    domain.PullRequestStateOpen,
			title:    "Add new feature",
			expected: utils.ColorGreenText,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := &integrations.IssueEvent{
				Issue: &github.Issue{
					State: github.String(tc.state),
					Title: github.String(tc.title),
				},
			}
			color := BuildIssueColor(payload)
			if color != tc.expected {
				t.Errorf("Expected color %s, got %s", tc.expected, color)
			}
		})
	}
}
