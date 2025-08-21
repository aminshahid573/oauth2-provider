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
	AuthCodeLifespan     = 10 * time.Minute
	RefreshTokenLifespan = 30 * 24 * time.Hour // 30 days
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

// GenerateAndStoreRefreshToken creates a new refresh token and stores its hash.
func (s *TokenService) GenerateAndStoreRefreshToken(ctx context.Context, userID, clientID string, scopes []string) (string, error) {
	// Generate a long, cryptographically secure random string for the token.
	token, err := utils.GenerateSecureToken(64)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	signature := hashToken(token)

	refreshToken := &models.Token{
		Signature: signature,
		ClientID:  clientID,
		UserID:    userID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(RefreshTokenLifespan),
		Type:      models.TokenTypeRefreshToken,
	}

	if err := s.tokenStore.Save(ctx, refreshToken); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

// ValidateAndConsumeRefreshToken checks if a refresh token is valid.
// Note: Refresh token rotation is a best practice, but for now, we'll just validate.
func (s *TokenService) ValidateAndConsumeRefreshToken(ctx context.Context, tokenStr string) (*models.Token, error) {
	signature := hashToken(tokenStr)

	token, err := s.tokenStore.GetBySignature(ctx, signature)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// In a simple implementation, we don't always delete the refresh token.
	// For refresh token rotation, you would delete the old one and issue a new one.
	// For now, we'll just check for expiration.

	if time.Now().After(token.ExpiresAt) {
		// If expired, we should delete it from the store.
		_ = s.tokenStore.DeleteBySignature(ctx, signature)
		return nil, fmt.Errorf("refresh token has expired")
	}

	if token.Type != models.TokenTypeRefreshToken {
		return nil, fmt.Errorf("invalid token type provided")
	}

	return token, nil
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
