package utils

import (
	"fmt"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomClaims defines the structure of our JWT claims.
type CustomClaims struct {
	Scope    []string `json:"scope,omitempty"`
	ClientID string   `json:"client_id"`
	jwt.RegisteredClaims
}

// JWTManager handles the creation and validation of JWTs.
type JWTManager struct {
	secretKey            []byte
	issuer               string
	accessTokenLifespan  time.Duration
	refreshTokenLifespan time.Duration
}

// NewJWTManager creates a new JWTManager.
func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey:            []byte(cfg.SecretKey),
		issuer:               cfg.Issuer,
		accessTokenLifespan:  cfg.AccessTokenLifespan,
		refreshTokenLifespan: cfg.RefreshTokenLifespan,
	}
}

// GenerateAccessToken creates a new JWT access token.
func (m *JWTManager) GenerateAccessToken(userID, clientID string, scopes []string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		Scope:    scopes,
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{clientID}, // Audience can be the client
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenLifespan)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return signedToken, nil
}

// VerifyToken parses and validates a token string.
func (m *JWTManager) VerifyToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
