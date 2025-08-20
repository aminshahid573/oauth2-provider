// File: internal/storage/mongodb/token.go
package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TokenRepository implements the storage.TokenStore interface for MongoDB.
type TokenRepository struct {
	collection *mongo.Collection
}

// NewTokenRepository creates a new TokenRepository.
func NewTokenRepository(db *mongo.Database) *TokenRepository {
	return &TokenRepository{
		collection: db.Collection("tokens"),
	}
}

// Save inserts a new token into the database.
func (r *TokenRepository) Save(ctx context.Context, token *models.Token) error {
	token.ID = bson.NewObjectID()
	token.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, token)
	return err
}

// GetBySignature retrieves a token by its signature.
func (r *TokenRepository) GetBySignature(ctx context.Context, signature string) (*models.Token, error) {
	var token models.Token
	filter := bson.M{"signature": signature}

	err := r.collection.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

// DeleteBySignature removes a token from the database by its signature.
func (r *TokenRepository) DeleteBySignature(ctx context.Context, signature string) error {
	filter := bson.M{"signature": signature}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}
