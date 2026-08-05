package issues

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/discord/colors"
)

type mockRepo struct {
	mappings map[string]*domain.MessageMapping
	getErr   error
	saveErr  error
	saved    *domain.MessageMapping
}

func (m *mockRepo) GetMapping(ctx context.Context, entityID string) (*domain.MessageMapping, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.mappings[entityID], nil
}

func (m *mockRepo) SaveMapping(ctx context.Context, mapping *domain.MessageMapping) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = mapping
	if m.mappings == nil {
		m.mappings = make(map[string]*domain.MessageMapping)
	}
	m.mappings[mapping.EntityID] = mapping
	return nil
}

func (m *mockRepo) DeleteMapping(ctx context.Context, entityID string) error {
	delete(m.mappings, entityID)
	return nil
}

type mockDiscord struct {
	sendMsgID string
	sendErr   error
	editErr   error

	lastChannelID string
	lastContent   string
	lastMessageID string
}

func (m *mockDiscord) SendMessage(ctx context.Context, channelID string, content string, embed interface{}) (string, error) {
	if m.sendErr != nil {
		return "", m.sendErr
	}
	m.lastChannelID = channelID
	m.lastContent = content
	return m.sendMsgID, nil
}

func (m *mockDiscord) EditMessage(ctx context.Context, channelID string, messageID string, embed interface{}) error {
	if m.editErr != nil {
		return m.editErr
	}
	m.lastChannelID = channelID
	m.lastMessageID = messageID
	return nil
}

func TestDetectType(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{name: "fix", title: "fix: resolve memory leak", expected: domain.TypeFix},
		{name: "hot uppercase", title: "HOT urgent issue", expected: domain.TypeHot},
		{name: "doc", title: "docs: update API documentation", expected: domain.TypeDoc},
		{name: "chore", title: "chore: update dependencies", expected: domain.TypeChore},
		{name: "other", title: "feature: add dark mode support", expected: domain.TypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectType(tt.title); got != tt.expected {
				t.Errorf("DetectType(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestBuildEmbed(t *testing.T) {
	tests := []struct {
		name          string
		state         string
		expectedColor int
	}{
		{name: "open issue", state: domain.IssueStateOpen, expectedColor: colors.Green},
		{name: "closed issue", state: domain.IssueStateClosed, expectedColor: colors.Grey},
		{name: "other state issue", state: "unknown", expectedColor: colors.Orange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &EventPayload{
				Issue: &github.Issue{
					Number:  github.Int(42),
					Title:   github.String("Test Issue"),
					Body:    github.String("Description of issue"),
					HTMLURL: github.String("https://github.com/org/repo/issues/42"),
					State:   github.String(tt.state),
				},
				Sender: &github.User{
					Login:     github.String("octocat"),
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

			if embed.Color != tt.expectedColor {
				t.Errorf("embed.Color = %d, want %d", embed.Color, tt.expectedColor)
			}
			if embed.Title != "Issue #42 - Test Issue" {
				t.Errorf("embed.Title = %q, want %q", embed.Title, "Issue #42 - Test Issue")
			}
			if embed.Author.Name != "octocat" {
				t.Errorf("embed.Author.Name = %q, want %q", embed.Author.Name, "octocat")
			}
		})
	}
}

func TestHandle(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repo GetMapping fails", func(t *testing.T) {
		repo := &mockRepo{getErr: errors.New("db error")}
		discord := &mockDiscord{}
		payload := &EventPayload{
			Issue: &github.Issue{ID: github.Int64(10)},
		}

		err := Handle(ctx, payload, repo, discord, "channel-1")
		if err == nil || err.Error() != "db error" {
			t.Errorf("Handle() error = %v, want 'db error'", err)
		}
	})

	t.Run("sends message and saves mapping for new issue", func(t *testing.T) {
		repo := &mockRepo{}
		discord := &mockDiscord{sendMsgID: "msg-123"}
		payload := &EventPayload{
			Issue: &github.Issue{
				ID:     github.Int64(100),
				Number: github.Int(1),
				Title:  github.String("New feature request"),
				State:  github.String(domain.IssueStateOpen),
			},
			Sender: &github.User{Login: github.String("user1")},
			Repository: &github.Repository{
				FullName: github.String("org/repo"),
				Owner:    &github.User{},
			},
		}

		err := Handle(ctx, payload, repo, discord, "channel-prs")
		if err != nil {
			t.Fatalf("Handle() unexpected error = %v", err)
		}

		if discord.lastContent != "" {
			t.Errorf("discord.lastContent = %q, want empty", discord.lastContent)
		}
		if repo.saved == nil || repo.saved.DiscordMessageID != "msg-123" {
			t.Errorf("repo.saved mapping = %v, want msgID msg-123", repo.saved)
		}
	})

	t.Run("mentions @everyone for new hot open issue", func(t *testing.T) {
		repo := &mockRepo{}
		discord := &mockDiscord{sendMsgID: "msg-456"}
		payload := &EventPayload{
			Issue: &github.Issue{
				ID:     github.Int64(200),
				Number: github.Int(2),
				Title:  github.String("HOT production outage"),
				State:  github.String(domain.IssueStateOpen),
			},
			Sender: &github.User{Login: github.String("user1")},
			Repository: &github.Repository{
				FullName: github.String("org/repo"),
				Owner:    &github.User{},
			},
		}

		err := Handle(ctx, payload, repo, discord, "channel-issues")
		if err != nil {
			t.Fatalf("Handle() unexpected error = %v", err)
		}

		if discord.lastContent != "@everyone" {
			t.Errorf("discord.lastContent = %q, want '@everyone'", discord.lastContent)
		}
	})

	t.Run("edits message for existing issue mapping", func(t *testing.T) {
		repo := &mockRepo{
			mappings: map[string]*domain.MessageMapping{
				"300": {
					EntityID:         "300",
					DiscordMessageID: "msg-789",
					DiscordChannelID: "channel-issues",
				},
			},
		}
		discord := &mockDiscord{}
		payload := &EventPayload{
			Issue: &github.Issue{
				ID:     github.Int64(300),
				Number: github.Int(3),
				Title:  github.String("Resolved issue"),
				State:  github.String(domain.IssueStateClosed),
			},
			Sender: &github.User{Login: github.String("user1")},
			Repository: &github.Repository{
				FullName: github.String("org/repo"),
				Owner:    &github.User{},
			},
		}

		err := Handle(ctx, payload, repo, discord, "channel-issues")
		if err != nil {
			t.Fatalf("Handle() unexpected error = %v", err)
		}

		if discord.lastMessageID != "msg-789" {
			t.Errorf("discord.lastMessageID = %q, want 'msg-789'", discord.lastMessageID)
		}
	})
}
