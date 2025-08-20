// File: internal/storage/interfaces.go
package storage

import (
	"context"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ClientStore defines the interface for client data storage.
type ClientStore interface {
	GetByClientID(ctx context.Context, clientID string) (*models.Client, error)
	Create(ctx context.Context, client *models.Client) error
	Update(ctx context.Context, client *models.Client) error
	Delete(ctx context.Context, clientID string) error
}

// UserStore defines the interface for user data storage.
type UserStore interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

// TokenStore defines the interface for token (auth code, refresh token) storage.
type TokenStore interface {
	Save(ctx context.Context, token *models.Token) error
	GetBySignature(ctx context.Context, signature string) (*models.Token, error)
	DeleteBySignature(ctx context.Context, signature string) error
}

// SessionStore defines the interface for user session storage (typically Redis).
type SessionStore interface {
	Save(ctx context.Context, session *models.Session) error
	Get(ctx context.Context, sessionID string) (*models.Session, error)
	Delete(ctx context.Context, sessionID string) error
}

// PKCEStore defines the interface for storing PKCE code challenges (typically Redis).
type PKCEStore interface {
	Save(ctx context.Context, code, challenge string, ttl time.Duration) error
	Get(ctx context.Context, code string) (string, error)
	Delete(ctx context.Context, code string) error
}

// DataStore is a composite interface that embeds all store interfaces.
// This is useful for dependency injection.
type DataStore struct {
	Client ClientStore
	User   UserStore
	Token  TokenStore
}
