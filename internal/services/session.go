package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const SessionLifespan = 24 * time.Hour // User login session duration

// SessionService provides logic for managing user sessions.
type SessionService struct {
	sessionStore storage.SessionStore
}

// NewSessionService creates a new SessionService.
func NewSessionService(sessionStore storage.SessionStore) *SessionService {
	return &SessionService{sessionStore: sessionStore}
}

// CreateSession creates a new login session for a user.
func (s *SessionService) CreateSession(ctx context.Context, userID bson.ObjectID) (*models.Session, error) {
	sessionID, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	session := &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(SessionLifespan),
	}

	if err := s.sessionStore.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// GetSession retrieves and validates a session by its ID.
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// DeleteSession ends a user's session (logout).
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.sessionStore.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
