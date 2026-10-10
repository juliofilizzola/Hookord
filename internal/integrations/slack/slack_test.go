package slack

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/common/utils"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
	slackLib "github.com/slack-go/slack"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockSlackClient struct {
	postChannel   string
	postTimestamp string
	postErr       error
	updateErr     error
	updateCalls   int
	postCalls     int
	newTimestamp  string
}

func (m *mockSlackClient) PostMessage(channelId string, _ slackLib.Attachment) (string, string, error) {
	m.postCalls++
	if m.postErr != nil {
		return "", "", m.postErr
	}
	return m.postChannel, m.postTimestamp, nil
}

func (m *mockSlackClient) UpdateMessage(channelId, ts string, _ slackLib.Attachment) (string, string, string, error) {
	m.updateCalls++
	if m.updateErr != nil {
		return "", "", "", m.updateErr
	}
	return channelId, m.newTimestamp, "", nil
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

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

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
			Owner: &github.User{
				Name:      github.String("org"),
				AvatarURL: github.String("https://example.com/owner.png"),
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Config / Validate
// ---------------------------------------------------------------------------

func TestConfig_Validate_MissingToken(t *testing.T) {
	cfg := Config{Token: ""}
	if err := cfg.Validate(); !errors.Is(err, ErrMissingToken) {
		t.Errorf("expected ErrMissingToken, got %v", err)
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := Config{Token: "xoxb-token", ChannelId: "C123"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Integration helpers (Name, Close, HandleIssue stub)
// ---------------------------------------------------------------------------

func TestIntegration_Name(t *testing.T) {
	i := New(Config{Token: "tok"}, &mockRepo{}, &mockSlackClient{})
	if got := i.Name(); got != "Slack" {
		t.Errorf("Name() = %q, want 'Slack'", got)
	}
}

func TestIntegration_Close(t *testing.T) {
	i := New(Config{Token: "tok"}, &mockRepo{}, &mockSlackClient{})
	if err := i.Close(); err != nil {
		t.Errorf("unexpected error on Close: %v", err)
	}
}

// ---------------------------------------------------------------------------
// New / NewWithToken
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	client := &mockSlackClient{}
	repo := &mockRepo{}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)
	if i == nil {
		t.Fatal("New returned nil")
	}
	if i.Name() != "Slack" {
		t.Errorf("expected Name 'Slack', got %q", i.Name())
	}
}

func TestNewWithToken_MissingToken(t *testing.T) {
	_, err := NewWithToken(Config{Token: ""}, &mockRepo{})
	if !errors.Is(err, ErrMissingToken) {
		t.Errorf("expected ErrMissingToken, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// HandlePullRequest
// ---------------------------------------------------------------------------

func TestHandlePullRequest_NilEvent(t *testing.T) {
	i := New(Config{Token: "tok", ChannelId: "C1"}, &mockRepo{}, &mockSlackClient{})

	if err := i.HandlePullRequest(context.Background(), nil); !errors.Is(err, ErrEventNil) {
		t.Errorf("expected ErrEventNil for nil event, got %v", err)
	}

	if err := i.HandlePullRequest(context.Background(), &integrations.PullRequestEvent{PullRequest: nil}); !errors.Is(err, ErrEventNil) {
		t.Errorf("expected ErrEventNil for nil PullRequest, got %v", err)
	}
}

func TestHandlePullRequest_NilClient(t *testing.T) {
	i := New(Config{Token: "tok", ChannelId: "C1"}, &mockRepo{}, nil)
	event := newPREvent(1, "feat: something", domain.PullRequestStateOpen)
	if err := i.HandlePullRequest(context.Background(), event); !errors.Is(err, ErrEventNil) {
		t.Errorf("expected ErrEventNil when client is nil, got %v", err)
	}
}

func TestHandlePullRequest_GetMappingError(t *testing.T) {
	repo := &mockRepo{getErr: errors.New("redis down")}
	client := &mockSlackClient{}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(1, "feat: test", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if err == nil || err.Error() != "redis down" {
		t.Errorf("expected 'redis down', got %v", err)
	}
}

func TestHandlePullRequest_NewMessage_Success(t *testing.T) {
	repo := &mockRepo{}
	client := &mockSlackClient{postChannel: "C1", postTimestamp: "ts-001"}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(42, "feat: add search", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.postCalls != 1 {
		t.Errorf("expected 1 PostMessage call, got %d", client.postCalls)
	}
	if repo.saved == nil {
		t.Fatal("expected mapping to be saved")
	}
	if repo.saved.SlackMessageID != "C1" {
		t.Errorf("SlackMessageID = %q, want 'C1'", repo.saved.SlackMessageID)
	}
	if repo.saved.SlackTimestamp != "ts-001" {
		t.Errorf("SlackTimestamp = %q, want 'ts-001'", repo.saved.SlackTimestamp)
	}
}

func TestHandlePullRequest_NewMessage_PostError(t *testing.T) {
	repo := &mockRepo{}
	client := &mockSlackClient{postErr: errors.New("slack rate limit")}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(1, "feat: search", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if err == nil || err.Error() != "slack rate limit" {
		t.Errorf("expected 'slack rate limit', got %v", err)
	}
}

func TestHandlePullRequest_NewMessage_SaveError(t *testing.T) {
	repo := &mockRepo{saveErr: errors.New("save failed")}
	client := &mockSlackClient{postChannel: "C1", postTimestamp: "ts-001"}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(1, "feat: search", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if err == nil || err.Error() != "save failed" {
		t.Errorf("expected 'save failed', got %v", err)
	}
}

func TestHandlePullRequest_UpdateMessage_Success(t *testing.T) {
	entityID := "99-Slack"
	repo := &mockRepo{
		mappings: map[string]*domain.MessageMapping{
			entityID: {
				EntityID:       entityID,
				SlackMessageID: "old-channel",
				SlackTimestamp: "ts-old",
			},
		},
	}
	client := &mockSlackClient{newTimestamp: "ts-new"}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(99, "fix: patch", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.updateCalls != 1 {
		t.Errorf("expected 1 UpdateMessage call, got %d", client.updateCalls)
	}
	if repo.saved.SlackTimestamp != "ts-new" {
		t.Errorf("SlackTimestamp = %q, want 'ts-new'", repo.saved.SlackTimestamp)
	}
}

func TestHandlePullRequest_UpdateMessage_EmptyTimestamp(t *testing.T) {
	entityID := "88-Slack"
	repo := &mockRepo{
		mappings: map[string]*domain.MessageMapping{
			entityID: {
				EntityID:       entityID,
				SlackMessageID: "",
				SlackTimestamp: "",
			},
		},
	}
	client := &mockSlackClient{}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(88, "feat: x", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if !errors.Is(err, ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestHandlePullRequest_UpdateMessage_UpdateError(t *testing.T) {
	entityID := "77-Slack"
	repo := &mockRepo{
		mappings: map[string]*domain.MessageMapping{
			entityID: {
				EntityID:       entityID,
				SlackMessageID: "C77",
				SlackTimestamp: "ts-77",
			},
		},
	}
	client := &mockSlackClient{updateErr: errors.New("update failed")}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(77, "doc: readme", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if err == nil || err.Error() != "update failed" {
		t.Errorf("expected 'update failed', got %v", err)
	}
}

func TestHandlePullRequest_UpdateMessage_SaveAfterUpdateError(t *testing.T) {
	entityID := "66-Slack"
	repo := &mockRepo{
		mappings: map[string]*domain.MessageMapping{
			entityID: {
				EntityID:       entityID,
				SlackMessageID: "C66",
				SlackTimestamp: "ts-66",
			},
		},
		saveErr: errors.New("save err after update"),
	}
	client := &mockSlackClient{newTimestamp: "ts-new"}
	i := New(Config{Token: "tok", ChannelId: "C1"}, repo, client)

	event := newPREvent(66, "fix: something", domain.PullRequestStateOpen)
	err := i.HandlePullRequest(context.Background(), event)
	if err == nil || err.Error() != "save err after update" {
		t.Errorf("expected 'save err after update', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// BuildPullRequestColor
// ---------------------------------------------------------------------------

func TestBuildPullRequestColor(t *testing.T) {
	tests := []struct {
		name   string
		draft  bool
		merged bool
		state  string
		title  string
		want   string
	}{
		{"draft", true, false, domain.PullRequestStateOpen, "", utils.ColorGreyText},
		{"merged", false, true, domain.PullRequestStateClosed, "", utils.ColorPurpleText},
		{"closed", false, false, domain.PullRequestStateClosed, "", utils.ColorDarkGreyText},
		{"fix", false, false, domain.PullRequestStateOpen, "fix: bug", utils.ColorOrangeText},
		{"hot", false, false, domain.PullRequestStateOpen, "hot: critical", utils.ColorRedText},
		{"doc", false, false, domain.PullRequestStateOpen, "doc: readme", utils.ColorBlueText},
		{"chore", false, false, domain.PullRequestStateOpen, "chore: deps", utils.ColorYellowText},
		{"feat/default", false, false, domain.PullRequestStateOpen, "feat: feature", utils.ColorGreenText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := newPREvent(1, tt.title, tt.state)
			event.PullRequest.Draft = github.Bool(tt.draft)
			event.PullRequest.Merged = github.Bool(tt.merged)

			got := BuildPullRequestColor(event)
			if got != tt.want {
				t.Errorf("BuildPullRequestColor() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BuildStatusPullRequest
// ---------------------------------------------------------------------------

func TestBuildStatusPullRequest(t *testing.T) {
	tests := []struct {
		name   string
		state  string
		draft  bool
		merged bool
		want   string
	}{
		{"open", domain.PullRequestStateOpen, false, false, domain.PullRequestStateOpen},
		{"draft", domain.PullRequestStateOpen, true, false, domain.PullRequestStateDraft},
		{"merged", domain.PullRequestStateClosed, false, true, domain.PullRequestStateMerged},
		{"closed", domain.PullRequestStateClosed, false, false, domain.PullRequestStateClosed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &github.PullRequest{
				State:  github.String(tt.state),
				Draft:  github.Bool(tt.draft),
				Merged: github.Bool(tt.merged),
			}
			obj := BuildStatusPullRequest(pr)
			if obj == nil {
				t.Fatal("BuildStatusPullRequest returned nil")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// escapeSlackMarkdownReservedCharacters
// ---------------------------------------------------------------------------

func TestEscapeSlackMarkdown(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello & world", "hello &amp; world"},
		{"a < b", "a &lt; b"},
		{"a > b", "a &gt; b"},
		{"a & b < c > d", "a &amp; b &lt; c &gt; d"},
		{"no special chars", "no special chars"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeSlackMarkdownReservedCharacters(tt.input)
			if got != tt.want {
				t.Errorf("escape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BuildPullRequestContent (TextBlockObject)
// ---------------------------------------------------------------------------

func TestBuildPullRequestContent_NilBranches(t *testing.T) {
	pr := &github.PullRequest{
		Number:  github.Int(1),
		Title:   github.String("feat: test"),
		HTMLURL: github.String("https://github.com/pr/1"),
		// Base and Head are nil
	}
	obj := BuildPullRequestContent(pr)
	if obj == nil {
		t.Fatal("BuildPullRequestContent returned nil")
	}
}

func TestBuildPullRequestContent_WithBranches(t *testing.T) {
	pr := &github.PullRequest{
		Number:  github.Int(2),
		Title:   github.String("fix: bug"),
		HTMLURL: github.String("https://github.com/pr/2"),
		Base:    &github.PullRequestBranch{Ref: github.String("main")},
		Head:    &github.PullRequestBranch{Ref: github.String("fix-branch")},
	}
	obj := BuildPullRequestContent(pr)
	if obj == nil {
		t.Fatal("BuildPullRequestContent returned nil")
	}
}

// ---------------------------------------------------------------------------
// BuildAssigneesPullRequest / BuildLabesPullRequest / BuildReviewsPullRequest
// ---------------------------------------------------------------------------

func TestBuildAssigneesPullRequest(t *testing.T) {
	pr := &github.PullRequest{
		Assignees: []*github.User{
			{Login: github.String("alice")},
			{Login: github.String("bob")},
		},
	}
	obj := BuildAssigneesPullRequest(pr)
	if obj == nil {
		t.Fatal("BuildAssigneesPullRequest returned nil")
	}
}

func TestBuildLabesPullRequest(t *testing.T) {
	pr := &github.PullRequest{
		Labels: []*github.Label{
			{Name: github.String("bug")},
			{Name: github.String("enhancement")},
		},
	}
	obj := BuildLabesPullRequest(pr)
	if obj == nil {
		t.Fatal("BuildLabesPullRequest returned nil")
	}
}

func TestBuildReviewsPullRequest(t *testing.T) {
	pr := &github.PullRequest{
		RequestedReviewers: []*github.User{
			{Login: github.String("carol")},
		},
	}
	obj := BuildReviewsPullRequest(pr)
	if obj == nil {
		t.Fatal("BuildReviewsPullRequest returned nil")
	}
}

// ---------------------------------------------------------------------------
// BuildStatsPullRequest
// ---------------------------------------------------------------------------

func TestBuildStatsPullRequest(t *testing.T) {
	pr := &github.PullRequest{
		ChangedFiles: github.Int(3),
		Additions:    github.Int(50),
		Deletions:    github.Int(10),
	}
	obj := BuildStatsPullRequest(pr)
	if obj == nil {
		t.Fatal("BuildStatsPullRequest returned nil")
	}
}

// ---------------------------------------------------------------------------
// BuildRepositoryPullRequest
// ---------------------------------------------------------------------------

func TestBuildRepositoryPullRequest(t *testing.T) {
	pr := &github.PullRequest{
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{FullName: github.String("org/repo")},
		},
	}
	obj := BuildRepositoryPullRequest(pr)
	if obj == nil {
		t.Fatal("BuildRepositoryPullRequest returned nil")
	}
}

// ---------------------------------------------------------------------------
// BuildBranchPullRequest
// ---------------------------------------------------------------------------

func TestBuildBranchPullRequest(t *testing.T) {
	t.Run("with branches", func(t *testing.T) {
		pr := &github.PullRequest{
			Base: &github.PullRequestBranch{Ref: github.String("main")},
			Head: &github.PullRequestBranch{Ref: github.String("dev")},
		}
		block := BuildBranchPullRequest(pr)
		if block == nil {
			t.Fatal("BuildBranchPullRequest returned nil")
		}
	})
	t.Run("nil branches", func(t *testing.T) {
		pr := &github.PullRequest{}
		block := BuildBranchPullRequest(pr)
		if block == nil {
			t.Fatal("BuildBranchPullRequest returned nil with nil branches")
		}
	})
}

// ---------------------------------------------------------------------------
// BuildFooterPullRequest
// ---------------------------------------------------------------------------

func TestBuildFooterPullRequest(t *testing.T) {
	block := BuildFooterPullRequest()
	if block == nil {
		t.Fatal("BuildFooterPullRequest returned nil")
	}
}

// ---------------------------------------------------------------------------
// BuildAuthorContextPullRequest
// ---------------------------------------------------------------------------

func TestBuildAuthorContextPullRequest(t *testing.T) {
	event := newPREvent(1, "feat: test", domain.PullRequestStateOpen)
	block := BuildAuthorContextPullRequest(event)
	if block == nil {
		t.Fatal("BuildAuthorContextPullRequest returned nil")
	}
}

// ---------------------------------------------------------------------------
// BuildPullRequestAttachment (integration)
// ---------------------------------------------------------------------------

func TestBuildPullRequestAttachment(t *testing.T) {
	event := newPREvent(1, "feat: test", domain.PullRequestStateOpen)
	att := BuildPullRequestAttachment(event)
	if len(att.Blocks.BlockSet) == 0 {
		t.Error("expected non-empty blocks in attachment")
	}
}

// ---------------------------------------------------------------------------
// BuildThumbnailAccessoryPullRequest / BuildPullRequestThumbnail
// ---------------------------------------------------------------------------

func TestBuildThumbnailAccessory(t *testing.T) {
	event := newPREvent(1, "feat: x", domain.PullRequestStateOpen)
	acc := BuildThumbnailAccessoryPullRequest(event)
	if acc == nil {
		t.Fatal("BuildThumbnailAccessoryPullRequest returned nil")
	}
}
