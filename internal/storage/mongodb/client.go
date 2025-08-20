// File: internal/storage/mongodb/client.go
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ClientRepository struct {
	collection *mongo.Collection
}

func NewClientRepository(db *mongo.Database) *ClientRepository {
	return &ClientRepository{
		collection: db.Collection("clients"),
	}
}

func (r *ClientRepository) GetByClientID(ctx context.Context, clientID string) (*models.Client, error) {
	var client models.Client
	filter := bson.M{"client_id": clientID}

	err := r.collection.FindOne(ctx, filter).Decode(&client)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, utils.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find client by id %s: %w", clientID, err)
	}
	return &client, nil
}

func (r *ClientRepository) Create(ctx context.Context, client *models.Client) error {
	client.ID = bson.NewObjectID()
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to create client with id %s: %w", client.ClientID, err)
	}
	return nil
}
