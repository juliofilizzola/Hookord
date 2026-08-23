package review

import (
	"context"
	"strconv"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

type Counter struct {
	repo domain.MessageRepository
}

func NewCounter(repo domain.MessageRepository) *Counter {
	return &Counter{
		repo: repo,
	}
}

// RecordReview increments the review count and tracks distinct reviewers for a Pull Request.
func (c *Counter) RecordReview(ctx context.Context, prID int64, repoFullName string, userLogin string) (*domain.MessageMapping, error) {
	if c == nil || c.repo == nil {
		return nil, nil
	}

	entityID := strconv.FormatInt(prID, 10)
	mapping, err := c.repo.GetMapping(ctx, entityID)
	if err != nil {
		return nil, err
	}

	if mapping == nil {
		mapping = &domain.MessageMapping{
			EntityID:   entityID,
			Repository: repoFullName,
			Reviewers:  make(map[string]bool),
		}
	}

	if mapping.Reviewers == nil {
		mapping.Reviewers = make(map[string]bool)
	}

	if userLogin != "" {
		mapping.TotalReviews++
		mapping.Reviewers[userLogin] = true
		mapping.TotalReviewers = len(mapping.Reviewers)
	}

	if err := c.repo.SaveMapping(ctx, mapping); err != nil {
		return nil, err
	}

	return mapping, nil
}

// GetReviewStats retrieves the current total reviews and unique reviewers count for a Pull Request.
func (c *Counter) GetReviewStats(ctx context.Context, prID int64) (totalReviews int, totalReviewers int, err error) {
	if c == nil || c.repo == nil {
		return 0, 0, nil
	}

	entityID := strconv.FormatInt(prID, 10)
	mapping, err := c.repo.GetMapping(ctx, entityID)
	if err != nil {
		return 0, 0, err
	}

	if mapping == nil {
		return 0, 0, nil
	}

	return mapping.TotalReviews, mapping.TotalReviewers, nil
}
