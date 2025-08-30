// File: internal/utils/jwt.go
package utils

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// CustomClaims defines the structure of our JWT claims.
type CustomClaims struct {
	Scope    []string `json:"scope,omitempty"`
	ClientID string   `json:"client_id"`
	jwt.RegisteredClaims
}

// JWTManager handles the creation and validation of JWTs.
type JWTManager struct {
	privateKey           *rsa.PrivateKey
	publicKey            *rsa.PublicKey
	keyID                string
	issuer               string
	accessTokenLifespan  time.Duration
	refreshTokenLifespan time.Duration
}

// NewJWTManager creates a new JWTManager, generating a new RSA key pair on startup.
func NewJWTManager(cfg config.JWTConfig) (*JWTManager, error) {
	pemData, err := base64.StdEncoding.DecodeString(cfg.PrivateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key from PEM: %w", err)
	}

	// Generate a unique ID for this key.
	keyID := uuid.NewString()

	slog.Info("new RSA key pair generated for JWT signing", "key_id", keyID)

	return &JWTManager{
		privateKey:           privateKey,
		publicKey:            &privateKey.PublicKey,
		keyID:                keyID,
		issuer:               cfg.Issuer,
		accessTokenLifespan:  cfg.AccessTokenLifespan,
		refreshTokenLifespan: cfg.RefreshTokenLifespan,
	}, nil
}

// GetPublicKeySet returns the public key as a JWK Set.
func (m *JWTManager) GetPublicKeySet() (jwk.Set, error) {
	key, err := jwk.FromRaw(m.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK from public key: %w", err)
	}

	key.Set(jwk.KeyIDKey, m.keyID)
	key.Set(jwk.AlgorithmKey, "RS256")
	key.Set(jwk.KeyUsageKey, jwk.ForSignature)

	keySet := jwk.NewSet()
	keySet.AddKey(key)
	return keySet, nil
}

// GenerateAccessToken creates a new JWT access token signed with the private key.
func (m *JWTManager) GenerateAccessToken(userID, clientID string, scopes []string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		Scope:    scopes,
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenLifespan)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	// Use RS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// Set the 'kid' (Key ID) in the header
	token.Header["kid"] = m.keyID

	signedToken, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return signedToken, nil
}

// VerifyToken parses and validates a token string using the public key.
func (m *JWTManager) VerifyToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect (RS256)
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the public key for verification
		return m.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetAccessTokenLifespan returns the configured lifespan for access tokens.
func (m *JWTManager) GetAccessTokenLifespan() time.Duration {
	return m.accessTokenLifespan
}
