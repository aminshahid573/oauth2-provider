package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"github.com/redis/go-redis/v9"
)

// PKCERepository implements the storage.PKCEStore interface for Redis.
type PKCERepository struct {
	client *redis.Client
}

// NewPKCERepository creates a new PKCERepository.
func NewPKCERepository(client *redis.Client) *PKCERepository {
	return &PKCERepository{client: client}
}

// Save stores a PKCE code challenge with a specific TTL.
func (r *PKCERepository) Save(ctx context.Context, code, challenge string, ttl time.Duration) error {
	if ttl <= 0 {
		return errors.New("pkce challenge ttl must be positive")
	}
	key := fmt.Sprintf("pkce:%s", code)
	if err := r.client.Set(ctx, key, challenge, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save pkce challenge to redis: %w", err)
	}
	return nil
}

// Get retrieves a PKCE code challenge.
func (r *PKCERepository) Get(ctx context.Context, code string) (string, error) {
	key := fmt.Sprintf("pkce:%s", code)
	challenge, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", utils.ErrNotFound
		}
		return "", fmt.Errorf("failed to get pkce challenge from redis: %w", err)
	}
	return challenge, nil
}

// Delete removes a PKCE code challenge after it has been used.
func (r *PKCERepository) Delete(ctx context.Context, code string) error {
	key := fmt.Sprintf("pkce:%s", code)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete pkce challenge from redis: %w", err)
	}
	return nil
}
