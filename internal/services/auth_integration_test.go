package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage/mongodb"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestAuthService_Integration tests the AuthService against a real MongoDB instance.
func TestAuthService_Integration(t *testing.T) {
	// Ensure we have a database connection.
	if testDB == nil {
		t.Skip("skipping integration test; no database connection")
	}

	// --- Setup for this specific test ---
	ctx := context.Background()
	userCollection := testDB.Collection("users")
	userStore := mongodb.NewUserRepository(testDB)
	authService := NewAuthService(userStore)

	// Clean up the collection before and after the test.
	userCollection.DeleteMany(ctx, bson.M{})
	defer userCollection.DeleteMany(ctx, bson.M{})

	// Create a test user.
	password := "strongpassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	testUser := &models.User{
		ID:             bson.NewObjectID(),
		Username:       "integ_test_user",
		HashedPassword: hashedPassword,
		Role:           "user",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = userCollection.InsertOne(ctx, testUser)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// --- Run Test Cases ---
	t.Run("Successful Authentication", func(t *testing.T) {
		user, err := authService.AuthenticateUser(ctx, testUser.Username, password)
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
		if user == nil {
			t.Fatal("expected user to be returned, but got nil")
		}
		if user.Username != testUser.Username {
			t.Errorf("expected username %s, but got %s", testUser.Username, user.Username)
		}
	})

	t.Run("Authentication with Wrong Password", func(t *testing.T) {
		_, err := authService.AuthenticateUser(ctx, testUser.Username, "wrongpassword")
		if err == nil {
			t.Error("expected an error, but got nil")
		}
		if !errors.Is(err, utils.ErrUnauthorized) {
			t.Errorf("expected error to be ErrUnauthorized, but got: %v", err)
		}
	})

	t.Run("Authentication for Non-existent User", func(t *testing.T) {
		_, err := authService.AuthenticateUser(ctx, "nonexistentuser", "anypassword")
		if err == nil {
			t.Error("expected an error, but got nil")
		}
		if !errors.Is(err, utils.ErrUnauthorized) {
			t.Errorf("expected error to be ErrUnauthorized, but got: %v", err)
		}
	})
}
