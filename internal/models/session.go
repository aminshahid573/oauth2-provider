package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Session represents a user's login session.
type Session struct {
	ID        string        `json:"id"`
	UserID    bson.ObjectID `json:"user_id"`
	ExpiresAt time.Time     `json:"expires_at"`
}
