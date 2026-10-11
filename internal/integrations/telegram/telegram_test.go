package telegram

import (
	"context"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

type mockTelegramClient struct {
	sendId    int
	sendErr   error
	editErr   error
	sendCalls int
	editCalls int
}

func (m *mockTelegramClient) SendMessage(_ string, _ string) (int, error) {
	m.sendCalls++
	if m.sendErr != nil {
		return 0, m.sendErr
	}
	return m.sendId, nil
}

func (m *mockTelegramClient) EditMessage(_ string, _ int, _ string) (int, error) {
	m.editCalls++
	if m.editErr != nil {
		return 0, m.editErr
	}
	return m.sendId, nil
}

func (m *mockTelegramClient) Close() error {
	return nil
}

type mockRepo struct {
	mappings map[string]*domain.MessageMapping
	getErr   error
	saveErr  error
	saved    *domain.MessageMapping
}

func (m *mockRepo) GetMapping(_ context.Context, entityID string) (*domain.MessageMapping, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.mappings == nil {
		return nil, nil
	}
	return m.mappings[entityID], nil
}

func (m *mockRepo) SaveMapping(_ context.Context, mapping *domain.MessageMapping) error {
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

func (m *mockRepo) DeleteMapping(_ context.Context, entityID string) error {
	if m.mappings != nil {
		delete(m.mappings, entityID)
	}
	return nil
}

func newPREvent(id int64, title, state string) *integrations.PullRequestEvent {
	return &integrations.PullRequestEvent{
		Action: "opened",
		PullRequest: &github.PullRequest{
			ID:     github.Int64(id),
			Number: github.Int(1),
			Title:  github.String(title),
			State:  github.String(state),
			User:   &github.User{Login: github.String("user"), AvatarURL: github.String("https://example.com/avatar.png")},
			Base: &github.PullRequestBranch{
				Ref:  github.String("main"),
				Repo: &github.Repository{FullName: github.String("org/repo")},
			},
			Head: &github.PullRequestBranch{Ref: github.String("feature")},
		},
		Sender: &github.User{Login: github.String("user"), AvatarURL: github.String("https://example.com/avatar.png")},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner:    &github.User{Name: github.String("org"), AvatarURL: github.String("https://example.com/org.png")},
		},
	}
}

func newIssueEvent(id int64, title, state string) *integrations.IssueEvent {
	return &integrations.IssueEvent{
		Action: "opened",
		Issue: &github.Issue{
			ID:     github.Int64(id),
			Number: github.Int(1),
			Title:  github.String(title),
			State:  github.String(state),
			User:   &github.User{Login: github.String("user"), AvatarURL: github.String("https://example.com/avatar.png"), Name: github.String("John")},
		},
		Sender: &github.User{Login: github.String("user"), AvatarURL: github.String("https://example.com/avatar.png")},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner:    &github.User{Name: github.String("org"), AvatarURL: github.String("https://example.com/org.png")},
		},
	}
}

func TestConfig_Validate_MissingToken(t *testing.T) {
	cfg := Config{}
	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected error for missing token")
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := Config{Token: "tok"}
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Unexpected error for valid config: %v", err)
	}
}

func TestIntegration_Name(t *testing.T) {
	i := New(Config{}, nil, nil)
	if i.Name() != Name {
		t.Errorf("Expected name %s, got %s", Name, i.Name())
	}
}

func TestIntegration_Close(t *testing.T) {
	i := New(Config{Token: "tok"}, &mockRepo{}, &mockTelegramClient{})
	if err := i.Close(); err != nil {
		t.Errorf("unexpected error on Close: %v", err)
	}
}

func TestNewWithToken_MissingToken(t *testing.T) {
	_, err := NewWithToken(Config{Token: ""}, &mockRepo{})
	if err == nil {
		t.Errorf("Expected error for missing token in NewWithToken")
	}
}
