package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/config"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

// generateSignature produces the X-Hub-Signature-256 header value expected by
// github.ValidatePayload.
func generateSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockRepo struct {
	mappings map[string]*domain.MessageMapping
	getErr   error
	saveErr  error
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

// mockIntegration satisfies integrations.Integration
type mockIntegration struct {
	name     string
	prErr    error
	issueErr error
}

func (m *mockIntegration) Name() string { return m.name }
func (m *mockIntegration) HandlePullRequest(_ context.Context, _ *integrations.PullRequestEvent) error {
	return m.prErr
}
func (m *mockIntegration) HandleIssue(_ context.Context, _ *integrations.IssueEvent) error {
	return m.issueErr
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newService(cfg *config.Config, repo domain.MessageRepository, i ...integrations.Integration) *WebhookService {
	manager := integrations.NewManager(i...)
	return NewWebhookService(cfg, repo, manager)
}

func validPRPayload(secret string) *http.Request {
	body := `{"action":"opened","pull_request":{"id":1,"number":1,"title":"feat: test","state":"open","base":{"ref":"main"},"head":{"ref":"feat"}},"repository":{"full_name":"org/repo"},"sender":{"login":"user"}}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")

	sig := generateSignature(secret, []byte(body))
	req.Header.Set("X-Hub-Signature-256", sig)
	return req
}

func validIssuePayload(secret string) *http.Request {
	body := `{"action":"opened","issue":{"id":99,"number":2,"title":"bug: crash","state":"open"},"repository":{"full_name":"org/repo"},"sender":{"login":"reporter"}}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "issues")

	sig := generateSignature(secret, []byte(body))
	req.Header.Set("X-Hub-Signature-256", sig)
	return req
}

func pingRequest() *http.Request {
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(`{}`))
	req.Header.Set("X-GitHub-Event", "ping")
	return req
}

func unknownEventRequest(secret string) *http.Request {
	body := `{"action":"pushed"}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "push")

	sig := generateSignature(secret, []byte(body))
	req.Header.Set("X-Hub-Signature-256", sig)
	return req
}

// ---------------------------------------------------------------------------
// NewWebhookService
// ---------------------------------------------------------------------------

func TestNewWebhookService(t *testing.T) {
	svc := newService(&config.Config{}, &mockRepo{})
	if svc == nil {
		t.Fatal("NewWebhookService returned nil")
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – ping
// ---------------------------------------------------------------------------

func TestHandleWebhook_Ping(t *testing.T) {
	svc := newService(&config.Config{GithubSecret: "secret"}, &mockRepo{})
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, pingRequest())

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 on ping, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – invalid signature (400)
// ---------------------------------------------------------------------------

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	svc := newService(&config.Config{GithubSecret: "correct-secret"}, &mockRepo{})

	body := `{"action":"opened"}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalidsignature")

	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on invalid signature, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – pull request (success)
// ---------------------------------------------------------------------------

func TestHandleWebhook_PullRequest_Success(t *testing.T) {
	const secret = "mysecret"
	svc := newService(&config.Config{GithubSecret: secret}, &mockRepo{}, &mockIntegration{name: "discord"})
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, validPRPayload(secret))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – pull request with nil config (uses default secret)
// ---------------------------------------------------------------------------

func TestHandleWebhook_NilConfig_UsesDefaultSecret(t *testing.T) {
	const defaultSecret = "your_github_webhook_secret"
	svc := newService(nil, &mockRepo{}, &mockIntegration{name: "discord"})
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, validPRPayload(defaultSecret))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – pull request (integration error → 500)
// ---------------------------------------------------------------------------

func TestHandleWebhook_PullRequest_IntegrationError(t *testing.T) {
	const secret = "mysecret"
	svc := newService(
		&config.Config{GithubSecret: secret},
		&mockRepo{},
		&mockIntegration{name: "discord", prErr: errors.New("discord down")},
	)
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, validPRPayload(secret))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – issue (success)
// ---------------------------------------------------------------------------

func TestHandleWebhook_Issue_Success(t *testing.T) {
	const secret = "mysecret"
	svc := newService(&config.Config{GithubSecret: secret}, &mockRepo{}, &mockIntegration{name: "discord"})
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, validIssuePayload(secret))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – issue (integration error → 500)
// ---------------------------------------------------------------------------

func TestHandleWebhook_Issue_IntegrationError(t *testing.T) {
	const secret = "mysecret"
	svc := newService(
		&config.Config{GithubSecret: secret},
		&mockRepo{},
		&mockIntegration{name: "discord", issueErr: errors.New("issue fail")},
	)
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, validIssuePayload(secret))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – unknown event type
// ---------------------------------------------------------------------------

func TestHandleWebhook_UnknownEvent(t *testing.T) {
	const secret = "mysecret"
	svc := newService(&config.Config{GithubSecret: secret}, &mockRepo{})
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, unknownEventRequest(secret))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for unhandled event, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook – empty GithubSecret uses default
// ---------------------------------------------------------------------------

func TestHandleWebhook_EmptySecret_FallsBackToDefault(t *testing.T) {
	const defaultSecret = "your_github_webhook_secret"
	svc := newService(&config.Config{GithubSecret: ""}, &mockRepo{})
	rr := httptest.NewRecorder()
	svc.HandleWebhook(rr, validPRPayload(defaultSecret))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with empty secret fallback, got %d", rr.Code)
	}
}
