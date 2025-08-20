package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents a resource owner in the system.
type User struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	Username       string        `bson:"username"`
	HashedPassword string        `bson:"hashed_password"`
	CreatedAt      time.Time     `bson:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at"`
}
