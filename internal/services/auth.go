package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// AuthService provides business logic for user authentication.
type AuthService struct {
	userStore storage.UserStore
}

// NewAuthService creates a new AuthService.
func NewAuthService(userStore storage.UserStore) *AuthService {
	return &AuthService{userStore: userStore}
}

// AuthenticateUser checks if a username and password combination is valid.
func (s *AuthService) AuthenticateUser(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.userStore.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			// Return a generic unauthorized error to prevent username enumeration.
			return nil, utils.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to get user for authentication: %w", err)
	}

	if !utils.CheckPasswordHash(password, user.HashedPassword) {
		return nil, utils.ErrUnauthorized
	}

	return user, nil
}
