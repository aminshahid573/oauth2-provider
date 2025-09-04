// File: internal/services/user.go (New File)
package services

import (
	"context"
	"fmt"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	Password string `json:"password" validate:"omitempty,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"omitempty,min=8"`
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

// GetUserByID retrieves a single user by their ID.
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	return s.userStore.GetByID(ctx, objID)
}

// UpdateUser handles the business logic for updating an existing user.
func (s *UserService) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) (*models.User, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	existingUser, err := s.userStore.GetByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	existingUser.Username = req.Username
	existingUser.Role = req.Role

	// Only update the password if a new one was provided.
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
		existingUser.HashedPassword = hashedPassword
	}

	if err := s.userStore.Update(ctx, existingUser); err != nil {
		return nil, err
	}
	return existingUser, nil
}

// DeleteUser deletes a user by their ID.
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}
	return s.userStore.Delete(ctx, objID)
}
