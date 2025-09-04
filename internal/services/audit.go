package services

import (
	"context"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
)

// AuditService provides a central way to record audit events.
type AuditService struct {
	store storage.AuditStore
}

// NewAuditService creates a new AuditService.
func NewAuditService(store storage.AuditStore) *AuditService {
	return &AuditService{store: store}
}

// RecordEventData holds the data needed to create an audit event.
type RecordEventData struct {
	EventType models.EventType
	ActorID   string
	TargetID  string
	IPAddress string
	UserAgent string
	Details   string
}

// Record creates and stores a new audit event from the provided data.
func (s *AuditService) Record(ctx context.Context, data RecordEventData) error {
	event := &models.AuditEvent{
		Timestamp: time.Now(),
		EventType: data.EventType,
		ActorID:   data.ActorID,
		TargetID:  data.TargetID,
		IPAddress: data.IPAddress,
		UserAgent: data.UserAgent,
		Details:   data.Details,
	}
	return s.store.Create(ctx, event)
}

// ListRecentEvents retrieves the most recent audit events.
func (s *AuditService) ListRecentEvents(ctx context.Context, limit int) ([]models.AuditEvent, error) {
	return s.store.ListRecent(ctx, limit)
}
