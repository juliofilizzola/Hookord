package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/juliofiliizzola/hookord/internal/application"
	"github.com/juliofiliizzola/hookord/internal/infrastructure/config"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestServer(t *testing.T) *Server {
	t.Helper()
	manager := integrations.NewManager()
	svc := application.NewWebhookService(&config.Config{}, nil, manager)
	return NewServer("0", svc) // port 0 = OS-assigned
}

// ---------------------------------------------------------------------------
// NewServer
// ---------------------------------------------------------------------------

func TestNewServer_NotNil(t *testing.T) {
	srv := newTestServer(t)
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

// ---------------------------------------------------------------------------
// /health endpoint
// ---------------------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	// Use the server's internal mux by extracting the handler via Shutdown trick:
	// Simpler: create a real httptest.Server from the internal handler.
	ts := httptest.NewServer(srv.srv.Handler)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health error: %v", err)
	}
	defer resp.Body.Close()
	_ = rr
	_ = req

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health = %d, want 200", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func TestServer_Shutdown(t *testing.T) {
	srv := newTestServer(t)

	// Start in background (port 0 → ephemeral port, just test Shutdown)
	go func() {
		// ListenAndServe on a real port would block; use :0 behaviour.
		// We call Start() in goroutine and then Shutdown immediately.
		_ = srv.Start()
	}()

	// Give Start() a moment to bind.
	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error: %v", err)
	}
}
