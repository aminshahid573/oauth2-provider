package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Constants for supported grant types.
const (
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeClientCredentials = "client_credentials"
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeDeviceCode        = "urn:ietf:params:oauth:grant-type:device_code"
	GrantTypeJWTBearer         = "urn:ietf:params:oauth:grant-type:jwt-bearer"
)

// Client represents an OAuth2 client application.
type Client struct {
	ID            bson.ObjectID `bson:"_id,omitempty"`
	ClientID      string        `bson:"client_id"`
	ClientSecret  string        `bson:"client_secret"` // This should be a hashed value
	Name          string        `bson:"name"`
	RedirectURIs  []string      `bson:"redirect_uris"`
	GrantTypes    []string      `bson:"grant_types"`
	ResponseTypes []string      `bson:"response_types"`
	Scopes        []string      `bson:"scopes"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
}
