package integrations

import (
	"context"
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// mock integration
// ---------------------------------------------------------------------------

type mockIntegration struct {
	name     string
	prErr    error
	issueErr error
	prCalls  int
	issCalls int
}

func (m *mockIntegration) Name() string { return m.name }

func (m *mockIntegration) HandlePullRequest(_ context.Context, _ *PullRequestEvent) error {
	m.prCalls++
	return m.prErr
}

func (m *mockIntegration) HandleIssue(_ context.Context, _ *IssueEvent) error {
	m.issCalls++
	return m.issueErr
}

// ---------------------------------------------------------------------------
// Manager tests
// ---------------------------------------------------------------------------

func TestNewManager_Empty(t *testing.T) {
	m := NewManager()
	if len(m.Integrations()) != 0 {
		t.Errorf("expected 0 integrations, got %d", len(m.Integrations()))
	}
}

func TestManager_Register(t *testing.T) {
	m := NewManager()
	i := &mockIntegration{name: "discord"}
	m.Register(i)

	if len(m.Integrations()) != 1 {
		t.Errorf("expected 1 integration after register, got %d", len(m.Integrations()))
	}
	if m.Integrations()[0].Name() != "discord" {
		t.Errorf("expected integration name 'discord', got %q", m.Integrations()[0].Name())
	}
}

func TestManager_HandlePullRequest_AllSuccess(t *testing.T) {
	i1 := &mockIntegration{name: "discord"}
	i2 := &mockIntegration{name: "slack"}
	m := NewManager(i1, i2)

	err := m.HandlePullRequest(context.Background(), &PullRequestEvent{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if i1.prCalls != 1 {
		t.Errorf("i1: expected 1 pr call, got %d", i1.prCalls)
	}
	if i2.prCalls != 1 {
		t.Errorf("i2: expected 1 pr call, got %d", i2.prCalls)
	}
}

func TestManager_HandlePullRequest_OneError(t *testing.T) {
	errDiscord := errors.New("discord failed")
	i1 := &mockIntegration{name: "discord", prErr: errDiscord}
	i2 := &mockIntegration{name: "slack"}
	m := NewManager(i1, i2)

	err := m.HandlePullRequest(context.Background(), &PullRequestEvent{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errDiscord) {
		t.Errorf("expected errDiscord in joined error, got %v", err)
	}
	// i2 must still be called (fan-out)
	if i2.prCalls != 1 {
		t.Errorf("i2 should still be called even if i1 failed, got %d calls", i2.prCalls)
	}
}

func TestManager_HandlePullRequest_AllErrors(t *testing.T) {
	e1 := errors.New("err1")
	e2 := errors.New("err2")
	i1 := &mockIntegration{name: "a", prErr: e1}
	i2 := &mockIntegration{name: "b", prErr: e2}
	m := NewManager(i1, i2)

	err := m.HandlePullRequest(context.Background(), &PullRequestEvent{})
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !errors.Is(err, e1) {
		t.Errorf("expected e1 in error, got %v", err)
	}
	if !errors.Is(err, e2) {
		t.Errorf("expected e2 in error, got %v", err)
	}
}

func TestManager_HandleIssue_AllSuccess(t *testing.T) {
	i1 := &mockIntegration{name: "discord"}
	i2 := &mockIntegration{name: "slack"}
	m := NewManager(i1, i2)

	err := m.HandleIssue(context.Background(), &IssueEvent{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if i1.issCalls != 1 {
		t.Errorf("i1: expected 1 issue call, got %d", i1.issCalls)
	}
	if i2.issCalls != 1 {
		t.Errorf("i2: expected 1 issue call, got %d", i2.issCalls)
	}
}

func TestManager_HandleIssue_OneError(t *testing.T) {
	errSlack := errors.New("slack failed")
	i1 := &mockIntegration{name: "discord"}
	i2 := &mockIntegration{name: "slack", issueErr: errSlack}
	m := NewManager(i1, i2)

	err := m.HandleIssue(context.Background(), &IssueEvent{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errSlack) {
		t.Errorf("expected errSlack in joined error, got %v", err)
	}
	if i1.issCalls != 1 {
		t.Errorf("i1 should still be called, got %d calls", i1.issCalls)
	}
}

func TestManager_HandlePullRequest_NoIntegrations(t *testing.T) {
	m := NewManager()
	err := m.HandlePullRequest(context.Background(), &PullRequestEvent{})
	if err != nil {
		t.Errorf("expected nil error with no integrations, got %v", err)
	}
}

func TestManager_HandleIssue_NoIntegrations(t *testing.T) {
	m := NewManager()
	err := m.HandleIssue(context.Background(), &IssueEvent{})
	if err != nil {
		t.Errorf("expected nil error with no integrations, got %v", err)
	}
}
