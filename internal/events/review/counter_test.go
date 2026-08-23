package review

import (
	"context"
	"errors"
	"testing"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

type mockRepo struct {
	mappings map[string]*domain.MessageMapping
	getErr   error
	saveErr  error
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

func TestRecordReview_FirstReview(t *testing.T) {
	repo := &mockRepo{}
	counter := NewCounter(repo)
	ctx := context.Background()

	mapping, err := counter.RecordReview(ctx, 31, "juliofilizzola/Hookord", "juliofilizzola")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mapping.TotalReviews != 1 {
		t.Errorf("TotalReviews = %d, want 1", mapping.TotalReviews)
	}
	if mapping.TotalReviewers != 1 {
		t.Errorf("TotalReviewers = %d, want 1", mapping.TotalReviewers)
	}
	if !mapping.Reviewers["juliofilizzola"] {
		t.Errorf("expected reviewer 'juliofilizzola' to be recorded")
	}
}

func TestRecordReview_MultipleReviewsSameUser(t *testing.T) {
	repo := &mockRepo{}
	counter := NewCounter(repo)
	ctx := context.Background()

	_, err := counter.RecordReview(ctx, 31, "juliofilizzola/Hookord", "juliofilizzola")
	if err != nil {
		t.Fatalf("unexpected error on 1st review: %v", err)
	}

	mapping, err := counter.RecordReview(ctx, 31, "juliofilizzola/Hookord", "juliofilizzola")
	if err != nil {
		t.Fatalf("unexpected error on 2nd review: %v", err)
	}

	if mapping.TotalReviews != 2 {
		t.Errorf("TotalReviews = %d, want 2", mapping.TotalReviews)
	}
	if mapping.TotalReviewers != 1 {
		t.Errorf("TotalReviewers = %d, want 1 (distinct user)", mapping.TotalReviewers)
	}
}

func TestRecordReview_MultipleUsers(t *testing.T) {
	repo := &mockRepo{}
	counter := NewCounter(repo)
	ctx := context.Background()

	_, _ = counter.RecordReview(ctx, 31, "juliofilizzola/Hookord", "user1")
	_, _ = counter.RecordReview(ctx, 31, "juliofilizzola/Hookord", "user1")
	mapping, err := counter.RecordReview(ctx, 31, "juliofilizzola/Hookord", "user2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mapping.TotalReviews != 3 {
		t.Errorf("TotalReviews = %d, want 3", mapping.TotalReviews)
	}
	if mapping.TotalReviewers != 2 {
		t.Errorf("TotalReviewers = %d, want 2", mapping.TotalReviewers)
	}
}

func TestGetReviewStats(t *testing.T) {
	repo := &mockRepo{}
	counter := NewCounter(repo)
	ctx := context.Background()

	total, reviewers, err := counter.GetReviewStats(ctx, 31)
	if err != nil {
		t.Fatalf("unexpected error on empty stats: %v", err)
	}
	if total != 0 || reviewers != 0 {
		t.Errorf("got (%d, %d), want (0, 0)", total, reviewers)
	}

	_, _ = counter.RecordReview(ctx, 31, "juliofilizzola/Hookord", "user1")

	total, reviewers, err = counter.GetReviewStats(ctx, 31)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || reviewers != 1 {
		t.Errorf("got (%d, %d), want (1, 1)", total, reviewers)
	}
}

func TestRecordReview_RepoErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("get error", func(t *testing.T) {
		repo := &mockRepo{getErr: errors.New("db error")}
		counter := NewCounter(repo)
		_, err := counter.RecordReview(ctx, 31, "org/repo", "user1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("save error", func(t *testing.T) {
		repo := &mockRepo{saveErr: errors.New("save error")}
		counter := NewCounter(repo)
		_, err := counter.RecordReview(ctx, 31, "org/repo", "user1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
