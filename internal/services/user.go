// File: internal/services/user.go (New File)
package services

import (
	"context"
	"fmt"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// UserService provides business logic for user management.
type UserService struct {
	userStore storage.UserStore
}

// NewUserService creates a new UserService.
func NewUserService(userStore storage.UserStore) *UserService {
	return &UserService{userStore: userStore}
}

// CreateUserRequest defines the payload for creating a new user.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

// CreateUser handles the business logic for creating a new user.
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		Role:           req.Role,
	}

	if err := s.userStore.Create(ctx, user); err != nil {
		// TODO: Handle duplicate username error specifically
		return nil, err
	}

	return user, nil
}

// ListUsers retrieves all users.
func (s *UserService) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.userStore.List(ctx)
}
