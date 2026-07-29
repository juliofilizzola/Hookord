package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/juliofiliizzola/hookord/internal/domain"
	redisClient "github.com/redis/go-redis/v9"
)

type repository struct {
	client *redisClient.Client
}

func NewRepository(url string) (domain.MessageRepository, error) {
	opts, err := redisClient.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redisClient.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &repository{client: client}, nil
}

func (r *repository) SaveMapping(ctx context.Context, mapping *domain.MessageMapping) error {
	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("mapping:%s", mapping.EntityID)
	return r.client.Set(ctx, key, data, 0).Err()
}

func (r *repository) GetMapping(ctx context.Context, entityID string) (*domain.MessageMapping, error) {
	key := fmt.Sprintf("mapping:%s", entityID)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redisClient.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var mapping domain.MessageMapping
	if err := json.Unmarshal([]byte(val), &mapping); err != nil {
		return nil, err
	}

	return &mapping, nil
}

func (r *repository) DeleteMapping(ctx context.Context, entityID string) error {
	key := fmt.Sprintf("mapping:%s", entityID)
	return r.client.Del(ctx, key).Err()
}
