package services

import (
	"context"
	"fmt"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// ClientService provides business logic for OAuth2 clients.
type ClientService struct {
	clientStore storage.ClientStore
	baseURL     string
}

// NewClientService creates a new ClientService.
func NewClientService(clientStore storage.ClientStore, baseURL string) *ClientService {
	return &ClientService{
		clientStore: clientStore,
		baseURL:     baseURL,
	}
}

func (s *ClientService) GetBaseURL() string {
	return s.baseURL
}

// ValidateClientCredentials checks if the provided client ID and secret are valid.
// It fetches the client and compares the hashed secret.
func (s *ClientService) ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.Client, error) {
	client, err := s.clientStore.GetByClientID(ctx, clientID)
	if err != nil {
		// If the client is not found, or any other error occurs, return an invalid client error.
		// This prevents leaking information about which client IDs exist.
		return nil, utils.ErrInvalidClient
	}

	// Compare the provided secret with the stored hash.
	if !utils.CheckPasswordHash(clientSecret, client.ClientSecret) {
		return nil, utils.ErrInvalidClient
	}

	return client, nil
}

// GetClient retrieves a client by its ID.
func (s *ClientService) GetClient(ctx context.Context, clientID string) (*models.Client, error) {
	client, err := s.clientStore.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return client, nil
}
