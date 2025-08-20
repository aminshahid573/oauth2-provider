package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

const (
	AuthCodeLifespan = 10 * time.Minute
)

// TokenService provides business logic for creating and managing tokens.
type TokenService struct {
	jwtManager *utils.JWTManager
	tokenStore storage.TokenStore
}

// NewTokenService creates a new TokenService.
func NewTokenService(jwtManager *utils.JWTManager, tokenStore storage.TokenStore) *TokenService {
	return &TokenService{
		jwtManager: jwtManager,
		tokenStore: tokenStore,
	}
}

// GenerateAccessToken creates a new JWT access token.
func (s *TokenService) GenerateAccessToken(userID, clientID string, scopes []string) (string, error) {
	return s.jwtManager.GenerateAccessToken(userID, clientID, scopes)
}

// GenerateAndStoreAuthorizationCode creates a new authorization code and stores its hash.
func (s *TokenService) GenerateAndStoreAuthorizationCode(ctx context.Context, userID, clientID string, scopes []string) (string, error) {
	// Generate a cryptographically secure random string for the code.
	code, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate authorization code: %w", err)
	}

	// Create a signature (hash) of the code to store in the database.
	// This is a security best practice: we don't store the raw code.
	signature := hashToken(code)

	token := &models.Token{
		Signature: signature,
		ClientID:  clientID,
		UserID:    userID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(AuthCodeLifespan),
		Type:      models.TokenTypeAuthorizationCode,
	}

	if err := s.tokenStore.Save(ctx, token); err != nil {
		return "", fmt.Errorf("failed to store authorization code: %w", err)
	}

	return code, nil
}

// ValidateAndConsumeAuthCode checks if an auth code is valid and deletes it.
func (s *TokenService) ValidateAndConsumeAuthCode(ctx context.Context, code string) (*models.Token, error) {
	signature := hashToken(code)

	token, err := s.tokenStore.GetBySignature(ctx, signature)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code: %w", err)
	}

	// IMPORTANT: Immediately delete the code to prevent reuse (as per OAuth2 spec).
	if err := s.tokenStore.DeleteBySignature(ctx, signature); err != nil {
		// Log this error but continue, as the code is now effectively invalid.
		// In a real system, you'd have more robust logging/alerting here.
		fmt.Printf("WARN: failed to delete consumed auth code: %v\n", err)
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("authorization code has expired")
	}

	if token.Type != models.TokenTypeAuthorizationCode {
		return nil, fmt.Errorf("invalid token type provided")
	}

	return token, nil
}

// hashToken creates a SHA-256 hash of a token string.
func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
