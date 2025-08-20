package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"github.com/redis/go-redis/v9"
)

// SessionRepository implements the storage.SessionStore interface for Redis.
type SessionRepository struct {
	client *redis.Client
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(client *redis.Client) *SessionRepository {
	return &SessionRepository{client: client}
}

// Save stores a user session in Redis with a TTL.
func (r *SessionRepository) Save(ctx context.Context, session *models.Session) error {
	key := fmt.Sprintf("session:%s", session.ID)
	ttl := time.Until(session.ExpiresAt)

	// Ensure we don't try to set a negative TTL
	if ttl <= 0 {
		return errors.New("session expiration must be in the future")
	}

	// Serialize the session struct to JSON
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Set the value in Redis
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save session to redis: %w", err)
	}
	return nil
}

// Get retrieves a user session from Redis by its ID.
func (r *SessionRepository) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, utils.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Delete removes a user session from Redis.
func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from redis: %w", err)
	}
	return nil
}
