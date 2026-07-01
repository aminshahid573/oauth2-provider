package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RevocationRepository implements the storage.RevocationStore interface
// using Redis. Each revoked JTI is stored as a key with a TTL matching
// the token's remaining lifetime, ensuring automatic cleanup.
type RevocationRepository struct {
	client *redis.Client
}

// NewRevocationRepository creates a new RevocationRepository.
func NewRevocationRepository(client *redis.Client) *RevocationRepository {
	return &RevocationRepository{client: client}
}

// Revoke marks a JTI as revoked. The entry expires after ttl, which should
// be set to the remaining lifetime of the access token.
func (r *RevocationRepository) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		// Token is already expired; nothing to revoke.
		return nil
	}
	key := fmt.Sprintf("revoked_jti:%s", jti)
	if err := r.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("failed to store revoked JTI in redis: %w", err)
	}
	return nil
}

// IsRevoked checks whether a JTI has been revoked.
func (r *RevocationRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("revoked_jti:%s", jti)
	_, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check revoked JTI in redis: %w", err)
	}
	return true, nil
}
