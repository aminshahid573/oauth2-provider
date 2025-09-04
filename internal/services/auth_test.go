package services

import (
	"context"
	"errors"
	"testing"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MockUserStore is a mock implementation of the storage.UserStore interface.
// It allows us to control its behavior for testing purposes.
type MockUserStore struct {
	GetByUsernameFunc func(ctx context.Context, username string) (*models.User, error)
	GetByIDFunc       func(ctx context.Context, id bson.ObjectID) (*models.User, error)
	CreateFunc        func(ctx context.Context, user *models.User) error
}

// GetByUsername calls the mock function.
func (m *MockUserStore) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username)
	}
	return nil, errors.New("GetByUsernameFunc not implemented")
}

// GetByID calls the mock function.
func (m *MockUserStore) GetByID(ctx context.Context, id bson.ObjectID) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

// Create calls the mock function.
func (m *MockUserStore) Create(ctx context.Context, user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return errors.New("CreateFunc not implemented")
}

// TestAuthService_Unit tests the AuthService in isolation using a mock store.
func TestAuthService_Unit(t *testing.T) {
	ctx := context.Background()
	password := "strongpassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Define a standard user for our tests.
	testUser := &models.User{
		Username:       "test_user",
		HashedPassword: hashedPassword,
	}

	t.Run("Successful Authentication", func(t *testing.T) {
		// Setup the mock to return our test user.
		mockStore := &MockUserStore{
			GetByUsernameFunc: func(ctx context.Context, username string) (*models.User, error) {
				if username == testUser.Username {
					return testUser, nil
				}
				return nil, utils.ErrNotFound
			},
		}
		authService := NewAuthService(mockStore)

		// Execute the method we are testing.
		user, err := authService.AuthenticateUser(ctx, testUser.Username, password)

		// Assert the results.
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
		if user == nil {
			t.Fatal("expected a user, but got nil")
		}
		if user.Username != testUser.Username {
			t.Errorf("expected username %s, but got %s", testUser.Username, user.Username)
		}
	})

	t.Run("Authentication with Wrong Password", func(t *testing.T) {
		// Setup the mock to return our test user.
		mockStore := &MockUserStore{
			GetByUsernameFunc: func(ctx context.Context, username string) (*models.User, error) {
				return testUser, nil
			},
		}
		authService := NewAuthService(mockStore)

		// Execute with the wrong password.
		_, err := authService.AuthenticateUser(ctx, testUser.Username, "wrongpassword")

		// Assert we get the correct error.
		if !errors.Is(err, utils.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, but got: %v", err)
		}
	})

	t.Run("Authentication for Non-existent User", func(t *testing.T) {
		// Setup the mock to return a "not found" error.
		mockStore := &MockUserStore{
			GetByUsernameFunc: func(ctx context.Context, username string) (*models.User, error) {
				return nil, utils.ErrNotFound // Simulate user not being in the DB
			},
		}
		authService := NewAuthService(mockStore)

		// Execute for a user that doesn't exist.
		_, err := authService.AuthenticateUser(ctx, "nonexistent", "anypassword")

		// Assert we get the correct error.
		if !errors.Is(err, utils.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, but got: %v", err)
		}
	})

	t.Run("Authentication with Database Error", func(t *testing.T) {
		// Setup the mock to return a generic database error.
		dbError := errors.New("database connection lost")
		mockStore := &MockUserStore{
			GetByUsernameFunc: func(ctx context.Context, username string) (*models.User, error) {
				return nil, dbError
			},
		}
		authService := NewAuthService(mockStore)

		// Execute the method.
		_, err := authService.AuthenticateUser(ctx, testUser.Username, password)

		// Assert that the original database error is preserved.
		if !errors.Is(err, dbError) {
			t.Errorf("expected the original db error to be wrapped, but it was not")
		}
	})
}
