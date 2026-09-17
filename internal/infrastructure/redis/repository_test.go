package redis

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/juliofiliizzola/hookord/internal/domain"
	redisClient "github.com/redis/go-redis/v9"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestRepository starts an in-memory Redis (miniredis) and returns a
// repository backed by it, plus the miniredis instance for inspection.
func newTestRepository(t *testing.T) (domain.MessageRepository, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	repo, err := NewRepository("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}
	return repo, mr
}

// ---------------------------------------------------------------------------
// NewRepository
// ---------------------------------------------------------------------------

func TestNewRepository_InvalidURL(t *testing.T) {
	_, err := NewRepository("not-a-valid-redis-url://???")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestNewRepository_UnreachableServer(t *testing.T) {
	// port 1 is reserved and should be unreachable
	_, err := NewRepository("redis://127.0.0.1:1")
	if err == nil {
		t.Error("expected error for unreachable server, got nil")
	}
}

// ---------------------------------------------------------------------------
// SaveMapping / GetMapping round-trip
// ---------------------------------------------------------------------------

func TestSaveAndGetMapping(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	mapping := &domain.MessageMapping{
		EntityID:         "pr-42",
		DiscordMessageID: "discord-msg-001",
		DiscordChannelID: "channel-001",
		TotalReviews:     3,
		TotalReviewers:   2,
		Reviewers:        map[string]bool{"alice": true, "bob": true},
	}

	if err := repo.SaveMapping(ctx, mapping); err != nil {
		t.Fatalf("SaveMapping error: %v", err)
	}

	got, err := repo.GetMapping(ctx, "pr-42")
	if err != nil {
		t.Fatalf("GetMapping error: %v", err)
	}
	if got == nil {
		t.Fatal("GetMapping returned nil for existing key")
	}
	if got.EntityID != mapping.EntityID {
		t.Errorf("EntityID = %q, want %q", got.EntityID, mapping.EntityID)
	}
	if got.DiscordMessageID != mapping.DiscordMessageID {
		t.Errorf("DiscordMessageID = %q, want %q", got.DiscordMessageID, mapping.DiscordMessageID)
	}
	if got.TotalReviews != mapping.TotalReviews {
		t.Errorf("TotalReviews = %d, want %d", got.TotalReviews, mapping.TotalReviews)
	}
}

// ---------------------------------------------------------------------------
// GetMapping – key not found
// ---------------------------------------------------------------------------

func TestGetMapping_NotFound(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	got, err := repo.GetMapping(ctx, "non-existent-id")
	if err != nil {
		t.Fatalf("GetMapping error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing key, got %+v", got)
	}
}

// ---------------------------------------------------------------------------
// GetMapping – corrupted JSON
// ---------------------------------------------------------------------------

func TestGetMapping_CorruptedJSON(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	// Manually inject invalid JSON directly into miniredis
	mr.Set("mapping:bad-key", "not-valid-json")

	_, err := repo.GetMapping(ctx, "bad-key")
	if err == nil {
		t.Error("expected JSON unmarshal error, got nil")
	}
}

// ---------------------------------------------------------------------------
// DeleteMapping
// ---------------------------------------------------------------------------

func TestDeleteMapping(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	mapping := &domain.MessageMapping{EntityID: "to-delete"}
	if err := repo.SaveMapping(ctx, mapping); err != nil {
		t.Fatalf("SaveMapping error: %v", err)
	}

	if err := repo.DeleteMapping(ctx, "to-delete"); err != nil {
		t.Fatalf("DeleteMapping error: %v", err)
	}

	got, err := repo.GetMapping(ctx, "to-delete")
	if err != nil {
		t.Fatalf("GetMapping after delete error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after delete, got %+v", got)
	}
}

// ---------------------------------------------------------------------------
// SaveMapping – JSON marshal error (unmarshalable field via interface{})
// Not easily triggered with the current struct, so we test the happy path
// with a struct that has all nullable fields set to zero values.
// ---------------------------------------------------------------------------

func TestSaveMapping_ZeroValues(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	mapping := &domain.MessageMapping{EntityID: "zero-val"}
	if err := repo.SaveMapping(ctx, mapping); err != nil {
		t.Fatalf("SaveMapping zero values error: %v", err)
	}

	got, err := repo.GetMapping(ctx, "zero-val")
	if err != nil {
		t.Fatalf("GetMapping error: %v", err)
	}
	if got == nil {
		t.Fatal("expected mapping, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetMapping – Redis generic error (simulated via closed miniredis)
// ---------------------------------------------------------------------------

func TestGetMapping_RedisError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	repo, err := NewRepository("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}

	// Close the miniredis server to simulate a connection error
	mr.Close()

	ctx := context.Background()
	_, err = repo.GetMapping(ctx, "any-key")
	if err == nil {
		t.Error("expected error after miniredis is closed, got nil")
	}
}

// ---------------------------------------------------------------------------
// DeleteMapping – error (simulated via closed miniredis)
// ---------------------------------------------------------------------------

func TestDeleteMapping_RedisError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	repo, err := NewRepository("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}

	mr.Close()

	ctx := context.Background()
	err = repo.DeleteMapping(ctx, "any-key")
	if err == nil {
		t.Error("expected error after miniredis is closed, got nil")
	}
}

// ---------------------------------------------------------------------------
// SaveMapping error – simulated via direct client with broken connection
// We bypass NewRepository to inject an unreachable client directly.
// ---------------------------------------------------------------------------

func TestSaveMapping_RedisError(t *testing.T) {
	// Create a client pointing to a closed server
	opts, _ := redisClient.ParseURL("redis://127.0.0.1:1")
	client := redisClient.NewClient(opts)

	r := &repository{client: client}
	ctx := context.Background()

	mapping := &domain.MessageMapping{EntityID: "fail"}
	err := r.SaveMapping(ctx, mapping)
	if err == nil {
		t.Error("expected error for unreachable redis, got nil")
	}
}

// ---------------------------------------------------------------------------
// JSON marshal: ensure all fields round-trip correctly
// ---------------------------------------------------------------------------

func TestMappingJSONRoundTrip(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	original := &domain.MessageMapping{
		EntityID:         "rt-1",
		DiscordMessageID: "dm-1",
		DiscordChannelID: "dc-1",
		SlackMessageID:   "sm-1",
		SlackTimestamp:   "ts-1",
		IntegrationName:  "discord",
		Repository:       "org/repo",
		LastStatus:       "open",
		TotalReviews:     5,
		TotalReviewers:   3,
		Reviewers:        map[string]bool{"alice": true},
	}

	if err := repo.SaveMapping(ctx, original); err != nil {
		t.Fatalf("SaveMapping error: %v", err)
	}

	got, err := repo.GetMapping(ctx, "rt-1")
	if err != nil {
		t.Fatalf("GetMapping error: %v", err)
	}

	// Compare via JSON for a clean deep-equal
	wantBytes, _ := json.Marshal(original)
	gotBytes, _ := json.Marshal(got)
	if string(wantBytes) != string(gotBytes) {
		t.Errorf("round-trip mismatch:\nwant: %s\n got: %s", wantBytes, gotBytes)
	}
}
