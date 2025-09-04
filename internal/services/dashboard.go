package services

import (
	"context"

	"github.com/aminshahid573/oauth2-provider/internal/storage"
)

// DashboardStats holds the aggregated statistics for the admin dashboard.
type DashboardStats struct {
	TotalClients int64 `json:"total_clients"`
	TotalUsers   int64 `json:"total_users"`
	ActiveTokens int64 `json:"active_tokens"`
}

// DashboardService provides business logic for the admin dashboard.
type DashboardService struct {
	clientStore storage.ClientStore
	userStore   storage.UserStore
	tokenStore  storage.TokenStore
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(clientStore storage.ClientStore, userStore storage.UserStore, tokenStore storage.TokenStore) *DashboardService {
	return &DashboardService{
		clientStore: clientStore,
		userStore:   userStore,
		tokenStore:  tokenStore,
	}
}

// GetStats gathers statistics from all data stores.
func (s *DashboardService) GetStats(ctx context.Context) (*DashboardStats, error) {
	// In a real high-traffic system, you might run these in parallel with a WaitGroup.
	// For now, sequential is fine.
	clientCount, err := s.clientStore.Count(ctx)
	if err != nil {
		return nil, err
	}

	userCount, err := s.userStore.Count(ctx)
	if err != nil {
		return nil, err
	}

	tokenCount, err := s.tokenStore.Count(ctx)
	if err != nil {
		return nil, err
	}

	stats := &DashboardStats{
		TotalClients: clientCount,
		TotalUsers:   userCount,
		ActiveTokens: tokenCount,
	}

	return stats, nil
}
