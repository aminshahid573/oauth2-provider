package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// EventType defines the type of an audit event.
type EventType string

// Constants for different audit event types.
const (
	UserLoginSuccess EventType = "USER_LOGIN_SUCCESS"
	TokenIssued      EventType = "TOKEN_ISSUED"
	ClientCreated    EventType = "CLIENT_CREATED"
	ClientDeleted    EventType = "CLIENT_DELETED"
	UserCreated      EventType = "USER_CREATED"
)

// AuditEvent represents a single logged action in the system.
type AuditEvent struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Timestamp time.Time     `bson:"timestamp"`
	EventType EventType     `bson:"event_type"`
	ActorID   string        `bson:"actor_id"`  // Who performed the action (e.g., User ID)
	TargetID  string        `bson:"target_id"` // What the action was performed on (e.g., Client ID)
	IPAddress string        `bson:"ip_address"`
	UserAgent string        `bson:"user_agent"`
	Details   string        `bson:"details"` // e.g., "Grant Type: client_credentials"
}
