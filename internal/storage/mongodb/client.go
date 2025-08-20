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

// ClientRepository implements the storage.ClientStore interface for MongoDB.
type ClientRepository struct {
	collection *mongo.Collection
}

// NewClientRepository creates a new ClientRepository.
func NewClientRepository(db *mongo.Database) *ClientRepository {
	return &ClientRepository{
		collection: db.Collection("clients"),
	}
}

// GetByClientID retrieves a client by its client_id.
func (r *ClientRepository) GetByClientID(ctx context.Context, clientID string) (*models.Client, error) {
	var client models.Client
	filter := bson.M{"client_id": clientID}

	err := r.collection.FindOne(ctx, filter).Decode(&client)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Use a standard error for not found cases
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &client, nil
}

// Create inserts a new client into the database.
func (r *ClientRepository) Create(ctx context.Context, client *models.Client) error {
	client.ID = bson.NewObjectID()
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, client)
	return err
}
