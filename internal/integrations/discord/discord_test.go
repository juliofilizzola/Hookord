package discord

import (
	"context"
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v60/github"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

// mockDiscordClient is an in-memory stub for DiscordClient without real network calls.
type mockDiscordClient struct {
	sendMsgID string
	sendErr   error
	editErr   error
	closeErr  error

	lastChannelID   string
	lastMessageSend *discordgo.MessageSend

	lastEditChannelID string
	lastEditMessageID string
	lastEditEmbed     *discordgo.MessageEmbed

	sendCalls  int
	editCalls  int
	closeCalls int
}

func (m *mockDiscordClient) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend) (*discordgo.Message, error) {
	m.sendCalls++
	m.lastChannelID = channelID
	m.lastMessageSend = data
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	return &discordgo.Message{ID: m.sendMsgID, ChannelID: channelID}, nil
}

func (m *mockDiscordClient) ChannelMessageEditEmbed(channelID string, messageID string, embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	m.editCalls++
	m.lastEditChannelID = channelID
	m.lastEditMessageID = messageID
	m.lastEditEmbed = embed
	if m.editErr != nil {
		return nil, m.editErr
	}
	return &discordgo.Message{ID: messageID, ChannelID: channelID}, nil
}

func (m *mockDiscordClient) Close() error {
	m.closeCalls++
	return m.closeErr
}

// mockRepo is an in-memory stub for domain.MessageRepository.
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
	if m.mappings == nil {
		return nil, nil
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
	if m.mappings != nil {
		delete(m.mappings, entityID)
	}
	return nil
}

func defaultTestConfig() Config {
	return Config{
		Token: "test-token",
		ChannelMappings: map[string]string{
			"pull_requests": "channel-pr-123",
			"issues":        "channel-issue-456",
		},
	}
}

// ---------------------- PULL REQUEST TESTS ----------------------

func TestHandlePullRequest_Success_NewMessage(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{sendMsgID: "msg-discord-001"}
	repo := &mockRepo{}
	cfg := defaultTestConfig()
	integration := New(cfg, repo, client)

	event := &integrations.PullRequestEvent{
		Action: "opened",
		PullRequest: &github.PullRequest{
			ID:        github.Int64(101),
			Number:    github.Int(1),
			Title:     github.String("feat: initial feature"),
			State:     github.String(domain.PullRequestStateOpen),
			HTMLURL:   github.String("https://github.com/org/repo/pull/1"),
			Additions: github.Int(50),
			Deletions: github.Int(10),
			Base:      &github.PullRequestBranch{Ref: github.String("main")},
			Head:      &github.PullRequestBranch{Ref: github.String("feature-1")},
		},
		Sender: &github.User{
			Login:     github.String("author1"),
			AvatarURL: github.String("https://example.com/avatar.png"),
		},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner: &github.User{
				AvatarURL: github.String("https://example.com/owner.png"),
			},
		},
		TotalReviews:   2,
		TotalReviewers: 1,
	}

	err := integration.HandlePullRequest(ctx, event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.sendCalls != 1 {
		t.Errorf("expected 1 send call, got %d", client.sendCalls)
	}
	if client.lastChannelID != "channel-pr-123" {
		t.Errorf("expected channel 'channel-pr-123', got %q", client.lastChannelID)
	}
	if client.lastMessageSend.Content != "Faça o CodeReview" {
		t.Errorf("expected content 'Faça o CodeReview', got %q", client.lastMessageSend.Content)
	}
	if client.lastMessageSend.Embed == nil {
		t.Fatal("expected embed to be non-nil")
	}
	if client.lastMessageSend.Embed.Color != ColorGreen {
		t.Errorf("expected color %d (Green), got %d", ColorGreen, client.lastMessageSend.Embed.Color)
	}

	// Verify repo saved mapping
	if repo.saved == nil {
		t.Fatal("expected mapping to be saved in repository")
	}
	if repo.saved.EntityID != "101" {
		t.Errorf("saved entityID = %q, want '101'", repo.saved.EntityID)
	}
	if repo.saved.DiscordMessageID != "msg-discord-001" {
		t.Errorf("saved messageID = %q, want 'msg-discord-001'", repo.saved.DiscordMessageID)
	}
	if repo.saved.DiscordChannelID != "channel-pr-123" {
		t.Errorf("saved channelID = %q, want 'channel-pr-123'", repo.saved.DiscordChannelID)
	}
	if repo.saved.TotalReviews != 2 {
		t.Errorf("saved TotalReviews = %d, want 2", repo.saved.TotalReviews)
	}
}

func TestHandlePullRequest_Success_EditMessage(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{}
	repo := &mockRepo{
		mappings: map[string]*domain.MessageMapping{
			"101": {
				EntityID:         "101",
				DiscordMessageID: "existing-msg-999",
				DiscordChannelID: "channel-pr-123",
				TotalReviews:     3,
				TotalReviewers:   2,
			},
		},
	}
	cfg := defaultTestConfig()
	integration := New(cfg, repo, client)

	event := &integrations.PullRequestEvent{
		Action: "synchronize",
		PullRequest: &github.PullRequest{
			ID:      github.Int64(101),
			Number:  github.Int(1),
			Title:   github.String("feat: initial feature updated"),
			State:   github.String(domain.PullRequestStateOpen),
			HTMLURL: github.String("https://github.com/org/repo/pull/1"),
			Base:    &github.PullRequestBranch{Ref: github.String("main")},
			Head:    &github.PullRequestBranch{Ref: github.String("feature-1")},
		},
		Sender: &github.User{Login: github.String("author1")},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner:    &github.User{},
		},
	}

	err := integration.HandlePullRequest(ctx, event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.sendCalls != 0 {
		t.Errorf("expected 0 send calls, got %d", client.sendCalls)
	}
	if client.editCalls != 1 {
		t.Errorf("expected 1 edit call, got %d", client.editCalls)
	}
	if client.lastEditMessageID != "existing-msg-999" {
		t.Errorf("lastEditMessageID = %q, want 'existing-msg-999'", client.lastEditMessageID)
	}
	if client.lastEditChannelID != "channel-pr-123" {
		t.Errorf("lastEditChannelID = %q, want 'channel-pr-123'", client.lastEditChannelID)
	}

	// Verify review stats were preserved from existing mapping
	if event.TotalReviews != 3 {
		t.Errorf("event.TotalReviews = %d, want 3", event.TotalReviews)
	}
	if event.TotalReviewers != 2 {
		t.Errorf("event.TotalReviewers = %d, want 2", event.TotalReviewers)
	}
}

// ---------------------- ISSUE TESTS ----------------------

func TestHandleIssue_Success_NewMessage(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{sendMsgID: "msg-issue-001"}
	repo := &mockRepo{}
	cfg := defaultTestConfig()
	integration := New(cfg, repo, client)

	event := &integrations.IssueEvent{
		Action: "opened",
		Issue: &github.Issue{
			ID:      github.Int64(202),
			Number:  github.Int(5),
			Title:   github.String("bug: crash on login"),
			Body:    github.String("Steps to reproduce..."),
			State:   github.String(domain.IssueStateOpen),
			HTMLURL: github.String("https://github.com/org/repo/issues/5"),
		},
		Sender: &github.User{
			Login:     github.String("reporter"),
			AvatarURL: github.String("https://example.com/avatar.png"),
		},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner: &github.User{
				AvatarURL: github.String("https://example.com/owner.png"),
			},
		},
	}

	err := integration.HandleIssue(ctx, event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.sendCalls != 1 {
		t.Errorf("expected 1 send call, got %d", client.sendCalls)
	}
	if client.lastChannelID != "channel-issue-456" {
		t.Errorf("expected channel 'channel-issue-456', got %q", client.lastChannelID)
	}
	if repo.saved == nil || repo.saved.DiscordMessageID != "msg-issue-001" {
		t.Errorf("saved mapping msgID = %v, want 'msg-issue-001'", repo.saved)
	}
}

func TestHandleIssue_Success_EditMessage(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{}
	repo := &mockRepo{
		mappings: map[string]*domain.MessageMapping{
			"202": {
				EntityID:         "202",
				DiscordMessageID: "msg-issue-existing",
				DiscordChannelID: "channel-issue-456",
			},
		},
	}
	cfg := defaultTestConfig()
	integration := New(cfg, repo, client)

	event := &integrations.IssueEvent{
		Action: "closed",
		Issue: &github.Issue{
			ID:     github.Int64(202),
			Number: github.Int(5),
			Title:  github.String("bug: crash on login"),
			State:  github.String(domain.IssueStateClosed),
		},
		Sender: &github.User{Login: github.String("reporter")},
		Repository: &github.Repository{
			FullName: github.String("org/repo"),
			Owner:    &github.User{},
		},
	}

	err := integration.HandleIssue(ctx, event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.sendCalls != 0 {
		t.Errorf("expected 0 send calls, got %d", client.sendCalls)
	}
	if client.editCalls != 1 {
		t.Errorf("expected 1 edit call, got %d", client.editCalls)
	}
	if client.lastEditMessageID != "msg-issue-existing" {
		t.Errorf("lastEditMessageID = %q, want 'msg-issue-existing'", client.lastEditMessageID)
	}
}

// ---------------------- BUSINESS RULES ----------------------

func TestBusinessRules_MentionsAndContent(t *testing.T) {
	t.Run("hot open PR triggers @everyone", func(t *testing.T) {
		pr := &github.PullRequest{
			Title: github.String("hot: urgent production issue"),
			State: github.String(domain.PullRequestStateOpen),
		}
		if got := BuildPullRequestContent(pr); got != "@everyone" {
			t.Errorf("BuildPullRequestContent() = %q, want '@everyone'", got)
		}
	})

	t.Run("hot closed PR does not trigger @everyone", func(t *testing.T) {
		pr := &github.PullRequest{
			Title: github.String("hot: urgent production issue"),
			State: github.String(domain.PullRequestStateClosed),
		}
		if got := BuildPullRequestContent(pr); got != "Faça o CodeReview" {
			t.Errorf("BuildPullRequestContent() = %q, want 'Faça o CodeReview'", got)
		}
	})

	t.Run("normal PR has default CodeReview content", func(t *testing.T) {
		pr := &github.PullRequest{
			Title: github.String("feat: add search"),
			State: github.String(domain.PullRequestStateOpen),
		}
		if got := BuildPullRequestContent(pr); got != "Faça o CodeReview" {
			t.Errorf("BuildPullRequestContent() = %q, want 'Faça o CodeReview'", got)
		}
	})

	t.Run("hot open Issue triggers @everyone", func(t *testing.T) {
		issue := &github.Issue{
			Title: github.String("HOT server crash"),
			State: github.String(domain.IssueStateOpen),
		}
		if got := BuildIssueContent(issue); got != "@everyone" {
			t.Errorf("BuildIssueContent() = %q, want '@everyone'", got)
		}
	})

	t.Run("normal open Issue has empty content", func(t *testing.T) {
		issue := &github.Issue{
			Title: github.String("bug: regular issue"),
			State: github.String(domain.IssueStateOpen),
		}
		if got := BuildIssueContent(issue); got != "" {
			t.Errorf("BuildIssueContent() = %q, want empty string", got)
		}
	})
}

func TestBusinessRules_Colors(t *testing.T) {
	tests := []struct {
		name          string
		state         string
		draft         bool
		merged        bool
		title         string
		expectedColor int
	}{
		{name: "draft PR", state: domain.PullRequestStateOpen, draft: true, expectedColor: ColorGrey},
		{name: "merged PR", state: domain.PullRequestStateClosed, merged: true, expectedColor: ColorPurple},
		{name: "closed unmerged PR", state: domain.PullRequestStateClosed, expectedColor: ColorDarkGrey},
		{name: "fix PR", state: domain.PullRequestStateOpen, title: "fix: resolve bug", expectedColor: ColorOrange},
		{name: "hot PR", state: domain.PullRequestStateOpen, title: "hot: critical fix", expectedColor: ColorRed},
		{name: "doc PR", state: domain.PullRequestStateOpen, title: "doc: update readme", expectedColor: ColorBlue},
		{name: "chore PR", state: domain.PullRequestStateOpen, title: "chore: update deps", expectedColor: ColorYellow},
		{name: "feat/default PR", state: domain.PullRequestStateOpen, title: "feat: add user page", expectedColor: ColorGreen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &integrations.PullRequestEvent{
				PullRequest: &github.PullRequest{
					State:  github.String(tt.state),
					Draft:  github.Bool(tt.draft),
					Merged: github.Bool(tt.merged),
					Title:  github.String(tt.title),
				},
				Repository: &github.Repository{Owner: &github.User{}},
				Sender:     &github.User{},
			}
			embed := BuildPullRequestEmbed(payload)
			if embed.Color != tt.expectedColor {
				t.Errorf("embed.Color = %d, want %d", embed.Color, tt.expectedColor)
			}
		})
	}
}

func TestBusinessRules_IssueColors(t *testing.T) {
	tests := []struct {
		name          string
		state         string
		expectedColor int
	}{
		{name: "open issue", state: domain.IssueStateOpen, expectedColor: ColorGreen},
		{name: "closed issue", state: domain.IssueStateClosed, expectedColor: ColorGrey},
		{name: "other issue state", state: "unknown", expectedColor: ColorOrange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &integrations.IssueEvent{
				Issue: &github.Issue{
					State: github.String(tt.state),
					Title: github.String("Test"),
				},
				Repository: &github.Repository{Owner: &github.User{}},
				Sender:     &github.User{},
			}
			embed := BuildIssueEmbed(payload)
			if embed.Color != tt.expectedColor {
				t.Errorf("embed.Color = %d, want %d", embed.Color, tt.expectedColor)
			}
		})
	}
}

// ---------------------- ERROR HANDLING ----------------------

func TestHandlePullRequest_DiscordSendError(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{sendErr: errors.New("discord rate limit")}
	repo := &mockRepo{}
	integration := New(defaultTestConfig(), repo, client)

	event := &integrations.PullRequestEvent{
		PullRequest: &github.PullRequest{
			ID:    github.Int64(10),
			State: github.String(domain.PullRequestStateOpen),
		},
		Sender:     &github.User{},
		Repository: &github.Repository{Owner: &github.User{}},
	}

	err := integration.HandlePullRequest(ctx, event)
	if err == nil || err.Error() != "discord rate limit" {
		t.Errorf("expected 'discord rate limit', got %v", err)
	}
	if repo.saved != nil {
		t.Error("expected mapping NOT to be saved on send failure")
	}
}

func TestHandlePullRequest_DiscordEditError(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{editErr: errors.New("discord message not found")}
	repo := &mockRepo{
		mappings: map[string]*domain.MessageMapping{
			"10": {
				EntityID:         "10",
				DiscordMessageID: "msg-old",
				DiscordChannelID: "channel-pr-123",
			},
		},
	}
	integration := New(defaultTestConfig(), repo, client)

	event := &integrations.PullRequestEvent{
		PullRequest: &github.PullRequest{
			ID:    github.Int64(10),
			State: github.String(domain.PullRequestStateOpen),
		},
		Sender:     &github.User{},
		Repository: &github.Repository{Owner: &github.User{}},
	}

	err := integration.HandlePullRequest(ctx, event)
	if err == nil || err.Error() != "discord message not found" {
		t.Errorf("expected 'discord message not found', got %v", err)
	}
}

func TestHandlePullRequest_RepoErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("get mapping error", func(t *testing.T) {
		repo := &mockRepo{getErr: errors.New("redis connection refused")}
		client := &mockDiscordClient{}
		integration := New(defaultTestConfig(), repo, client)

		event := &integrations.PullRequestEvent{
			PullRequest: &github.PullRequest{ID: github.Int64(1)},
		}

		err := integration.HandlePullRequest(ctx, event)
		if err == nil || err.Error() != "redis connection refused" {
			t.Errorf("expected 'redis connection refused', got %v", err)
		}
		if client.sendCalls != 0 {
			t.Error("client should not be called if repo fails")
		}
	})

	t.Run("save mapping error", func(t *testing.T) {
		repo := &mockRepo{saveErr: errors.New("redis save failed")}
		client := &mockDiscordClient{sendMsgID: "msg-123"}
		integration := New(defaultTestConfig(), repo, client)

		event := &integrations.PullRequestEvent{
			PullRequest: &github.PullRequest{
				ID:    github.Int64(1),
				State: github.String(domain.PullRequestStateOpen),
			},
			Sender:     &github.User{},
			Repository: &github.Repository{Owner: &github.User{}},
		}

		err := integration.HandlePullRequest(ctx, event)
		if err == nil || err.Error() != "redis save failed" {
			t.Errorf("expected 'redis save failed', got %v", err)
		}
	})
}

// ---------------------- INTEGRATION UNAVAILABLE / UNCONFIGURED ----------------------

func TestHandlePullRequest_ClientUnavailable(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}
	integration := New(defaultTestConfig(), repo, nil) // nil client

	event := &integrations.PullRequestEvent{
		PullRequest: &github.PullRequest{ID: github.Int64(1)},
	}

	err := integration.HandlePullRequest(ctx, event)
	if !errors.Is(err, ErrClientUnavailable) {
		t.Errorf("expected ErrClientUnavailable, got %v", err)
	}
}

func TestHandlePullRequest_ChannelNotConfigured(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{}
	repo := &mockRepo{}
	cfg := Config{Token: "test", ChannelMappings: map[string]string{}} // empty channel mapping
	integration := New(cfg, repo, client)

	event := &integrations.PullRequestEvent{
		PullRequest: &github.PullRequest{ID: github.Int64(1)},
	}

	err := integration.HandlePullRequest(ctx, event)
	if !errors.Is(err, ErrChannelNotConfigured) {
		t.Errorf("expected ErrChannelNotConfigured, got %v", err)
	}
}

func TestHandleIssue_ChannelNotConfigured(t *testing.T) {
	ctx := context.Background()
	client := &mockDiscordClient{}
	repo := &mockRepo{}
	cfg := Config{Token: "test", ChannelMappings: map[string]string{}}
	integration := New(cfg, repo, client)

	event := &integrations.IssueEvent{
		Issue: &github.Issue{ID: github.Int64(1)},
	}

	err := integration.HandleIssue(ctx, event)
	if !errors.Is(err, ErrChannelNotConfigured) {
		t.Errorf("expected ErrChannelNotConfigured, got %v", err)
	}
}

func TestHandlePullRequest_NilEvent(t *testing.T) {
	ctx := context.Background()
	integration := New(defaultTestConfig(), &mockRepo{}, &mockDiscordClient{})

	if err := integration.HandlePullRequest(ctx, nil); !errors.Is(err, ErrNilEvent) {
		t.Errorf("expected ErrNilEvent, got %v", err)
	}

	if err := integration.HandlePullRequest(ctx, &integrations.PullRequestEvent{PullRequest: nil}); !errors.Is(err, ErrNilEvent) {
		t.Errorf("expected ErrNilEvent, got %v", err)
	}
}

func TestHandleIssue_NilEvent(t *testing.T) {
	ctx := context.Background()
	integration := New(defaultTestConfig(), &mockRepo{}, &mockDiscordClient{})

	if err := integration.HandleIssue(ctx, nil); !errors.Is(err, ErrNilEvent) {
		t.Errorf("expected ErrNilEvent, got %v", err)
	}
}

func TestIntegration_NameAndClose(t *testing.T) {
	client := &mockDiscordClient{}
	integration := New(defaultTestConfig(), &mockRepo{}, client)

	if got := integration.Name(); got != "discord" {
		t.Errorf("integration.Name() = %q, want 'discord'", got)
	}

	if err := integration.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
	if client.closeCalls != 1 {
		t.Errorf("expected 1 close call, got %d", client.closeCalls)
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{Token: "bot-token"}
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		cfg := Config{Token: ""}
		if err := cfg.Validate(); !errors.Is(err, ErrMissingToken) {
			t.Errorf("expected ErrMissingToken, got %v", err)
		}
	})
}
