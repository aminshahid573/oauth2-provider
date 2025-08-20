package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// TokenType defines the type of the stored token.
type TokenType string

const (
	TokenTypeAuthorizationCode TokenType = "auth_code"
	TokenTypeRefreshToken      TokenType = "refresh_token"
	TokenTypeDeviceCode        TokenType = "device_code"
)

// Token represents a stored authorization code or refresh token.
// We store a signature/hash of the token/code for security, not the raw value.
type Token struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Signature string        `bson:"signature"` // A hash of the actual token/code
	ClientID  string        `bson:"client_id"`
	UserID    string        `bson:"user_id"`
	Scopes    []string      `bson:"scopes"`
	ExpiresAt time.Time     `bson:"expires_at"`
	Type      TokenType     `bson:"type"`
	CreatedAt time.Time     `bson:"created_at"`
}
