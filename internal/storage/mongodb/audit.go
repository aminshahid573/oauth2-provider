package mongodb

import (
	"context"
	"fmt"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AuditRepository implements the storage.AuditStore interface.
type AuditRepository struct {
	collection *mongo.Collection
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(db *mongo.Database) *AuditRepository {
	return &AuditRepository{
		collection: db.Collection("audit_events"),
	}
}

// Create inserts a new audit event into the database.
func (r *AuditRepository) Create(ctx context.Context, event *models.AuditEvent) error {
	_, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to create audit event: %w", err)
	}
	return nil
}

// ListRecent retrieves the most recent audit events.
func (r *AuditRepository) ListRecent(ctx context.Context, limit int) ([]models.AuditEvent, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent audit events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []models.AuditEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode audit events: %w", err)
	}
	return events, nil
}
