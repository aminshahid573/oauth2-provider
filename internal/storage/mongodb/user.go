package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UserRepository implements the storage.UserStore interface for MongoDB.
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// GetByUsername retrieves a user by their username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	filter := bson.M{"username": username}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by their BSON ObjectID.
func (r *UserRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.User, error) {
	var user models.User
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = bson.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	return err
}
