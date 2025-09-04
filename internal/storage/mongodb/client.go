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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

// List retrieves all clients from the database.
func (r *ClientRepository) List(ctx context.Context) ([]models.Client, error) {
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find clients: %w", err)
	}
	defer cursor.Close(ctx)

	var clients []models.Client
	if err := cursor.All(ctx, &clients); err != nil {
		return nil, fmt.Errorf("failed to decode clients: %w", err)
	}
	return clients, nil
}

// Update replaces an existing client document.
func (r *ClientRepository) Update(ctx context.Context, client *models.Client) error {
	client.UpdatedAt = time.Now()
	filter := bson.M{"client_id": client.ClientID}

	result, err := r.collection.ReplaceOne(ctx, filter, client)
	if err != nil {
		return fmt.Errorf("failed to update client %s: %w", client.ClientID, err)
	}
	if result.MatchedCount == 0 {
		return utils.ErrNotFound
	}
	return nil
}

// Delete removes a client from the database by its client_id.
func (r *ClientRepository) Delete(ctx context.Context, clientID string) error {
	filter := bson.M{"client_id": clientID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete client %s: %w", clientID, err)
	}
	if result.DeletedCount == 0 {
		return utils.ErrNotFound
	}
	return nil
}

// Count returns the total number of client documents.
func (r *ClientRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count clients: %w", err)
	}
	return count, nil
}
