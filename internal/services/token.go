// File: internal/services/token.go
package services

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// TODO:: load from config , already defined in config
const (
	AuthCodeLifespan      = 10 * time.Minute
	RefreshTokenLifespan  = 30 * 24 * time.Hour // 30 days
	DeviceCodeLifespan    = 15 * time.Minute
	PKCEChallengeLifespan = 10 * time.Minute
)

// TokenService provides business logic for creating and managing tokens.
type TokenService struct {
	jwtManager *utils.JWTManager
	tokenStore storage.TokenStore
	pkceStore  storage.PKCEStore
}

// NewTokenService creates a new TokenService.
func NewTokenService(jwtManager *utils.JWTManager, tokenStore storage.TokenStore, pkceStore storage.PKCEStore) *TokenService {
	return &TokenService{
		jwtManager: jwtManager,
		tokenStore: tokenStore,
		pkceStore:  pkceStore,
	}
}

// --- Public Helper Methods ---

// HashToken creates a SHA-256 hash of a token string.
func (s *TokenService) HashToken(token string) string {
	return hashToken(token)
}

// GetAccessTokenLifespan returns the configured lifespan for access tokens.
func (s *TokenService) GetAccessTokenLifespan() time.Duration {
	return s.jwtManager.GetAccessTokenLifespan() // Assumes JWTManager exposes this
}

// --- Token Generation ---

// GenerateAccessToken creates a new JWT access token.
func (s *TokenService) GenerateAccessToken(userID, clientID string, scopes []string) (string, error) {
	return s.jwtManager.GenerateAccessToken(userID, clientID, scopes)
}

// GenerateAndStoreAuthorizationCode creates a new authorization code and stores its hash.
func (s *TokenService) GenerateAndStoreAuthorizationCode(ctx context.Context, userID, clientID string, scopes []string) (string, error) {
	code, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate authorization code: %w", err)
	}
	token := &models.Token{
		Signature: hashToken(code),
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
	token, err := utils.GenerateSecureToken(64)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken := &models.Token{
		Signature: hashToken(token),
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

// GenerateAndStoreDeviceCode creates a new device and user code.
func (s *TokenService) GenerateAndStoreDeviceCode(ctx context.Context, clientID string, scopes []string) (string, string, error) {
	deviceCode, err := utils.GenerateSecureToken(64)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate device code: %w", err)
	}
	userCode, err := utils.GenerateSecureToken(8)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate user code: %w", err)
	}
	token := &models.Token{
		Signature: hashToken(deviceCode),
		UserCode:  strings.ToUpper(userCode),
		ClientID:  clientID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(DeviceCodeLifespan),
		Type:      models.TokenTypeDeviceCode,
		Approved:  false,
	}
	if err := s.tokenStore.Save(ctx, token); err != nil {
		return "", "", fmt.Errorf("failed to store device code: %w", err)
	}
	return deviceCode, token.UserCode, nil
}

// --- Token Validation and Retrieval ---

// GetTokenBySignature retrieves a token by its signature.
func (s *TokenService) GetTokenBySignature(ctx context.Context, signature string) (*models.Token, error) {
	return s.tokenStore.GetBySignature(ctx, signature)
}

// GetTokenByUserCode retrieves a device code token by its user-facing code.
func (s *TokenService) GetTokenByUserCode(ctx context.Context, userCode string) (*models.Token, error) {
	return s.tokenStore.GetByUserCode(ctx, strings.ToUpper(userCode))
}

// ValidateAndConsumeAuthCode checks if an auth code is valid and deletes it.
func (s *TokenService) ValidateAndConsumeAuthCode(ctx context.Context, code string) (*models.Token, error) {
	signature := hashToken(code)
	token, err := s.tokenStore.GetBySignature(ctx, signature)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code: %w", err)
	}
	if err := s.tokenStore.DeleteBySignature(ctx, signature); err != nil {
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

// ValidateAndConsumeRefreshToken checks if a refresh token is valid.
func (s *TokenService) ValidateAndConsumeRefreshToken(ctx context.Context, tokenStr string) (*models.Token, error) {
	signature := hashToken(tokenStr)
	token, err := s.tokenStore.GetBySignature(ctx, signature)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	if time.Now().After(token.ExpiresAt) {
		_ = s.tokenStore.DeleteBySignature(ctx, signature)
		return nil, fmt.Errorf("refresh token has expired")
	}
	if token.Type != models.TokenTypeRefreshToken {
		return nil, fmt.Errorf("invalid token type provided")
	}
	return token, nil
}

// ApproveDeviceCode marks a device code as approved by a user.
func (s *TokenService) ApproveDeviceCode(ctx context.Context, userCode, userID string) (*models.Token, error) {
	token, err := s.GetTokenByUserCode(ctx, userCode)
	if err != nil {
		return nil, err
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("device code has expired")
	}
	token.Approved = true
	token.UserID = userID
	if err := s.tokenStore.Update(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

// DeleteTokenBySignature deletes a token by its signature.
func (s *TokenService) DeleteTokenBySignature(ctx context.Context, signature string) error {
	return s.tokenStore.DeleteBySignature(ctx, signature)
}

// StorePKCEChallenge saves the code challenge associated with an authorization code.
func (s *TokenService) StorePKCEChallenge(ctx context.Context, code, challenge string) error {
	return s.pkceStore.Save(ctx, code, challenge, PKCEChallengeLifespan)
}

// ValidatePKCE retrieves the stored challenge for a code and validates it against the verifier.
func (s *TokenService) ValidatePKCE(ctx context.Context, code, verifier string) error {
	challenge, err := s.pkceStore.Get(ctx, code)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			// If no challenge is found, it might be a flow that didn't use PKCE,
			// or the code is simply invalid. For security, we treat it as an error.
			return fmt.Errorf("invalid authorization code or no PKCE challenge found")
		}
		return fmt.Errorf("failed to retrieve PKCE challenge: %w", err)
	}

	// IMPORTANT: Delete the challenge immediately after retrieval to prevent reuse.
	defer s.pkceStore.Delete(ctx, code)

	// Check if the client provided a verifier. If a challenge was stored, a verifier is mandatory.
	if verifier == "" {
		return fmt.Errorf("code_verifier is required")
	}

	// Calculate the challenge from the provided verifier.
	generatedChallenge := utils.GeneratePKCEChallengeS256(verifier)

	// Compare in constant time to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(generatedChallenge)) != 1 {
		return fmt.Errorf("invalid code_verifier")
	}

	return nil
}

// --- Private Helpers ---

// hashToken creates a SHA-256 hash of a token string.
func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
