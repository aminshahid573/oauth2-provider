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

// CreateClientRequest defines the payload for creating a new client.
type CreateClientRequest struct {
	Name          string   `json:"name" validate:"required"`
	RedirectURIs  []string `json:"redirect_uris" validate:"required,dive,url"`
	GrantTypes    []string `json:"grant_types" validate:"required,dive,oneof=authorization_code client_credentials refresh_token urn:ietf:params:oauth:grant-type:device_code urn:ietf:params:oauth:grant-type:jwt-bearer"`
	ResponseTypes []string `json:"response_types" validate:"required"`
	Scopes        []string `json:"scopes" validate:"required"`
	JWKSURL       string   `json:"jwks_url" validate:"omitempty,url"`
}

type UpdateClientRequest struct {
	Name          string   `json:"name" validate:"required"`
	RedirectURIs  []string `json:"redirect_uris" validate:"required,dive,url"`
	GrantTypes    []string `json:"grant_types" validate:"required,dive,oneof=authorization_code client_credentials refresh_token urn:ietf:params:oauth:grant-type:device_code urn:ietf:params:oauth:grant-type:jwt-bearer"`
	ResponseTypes []string `json:"response_types" validate:"required"`
	Scopes        []string `json:"scopes" validate:"required"`
	JWKSURL       string   `json:"jwks_url" validate:"omitempty,url"`
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

// CreateClient handles the business logic for creating a new client.
// It returns the client with the plaintext secret for one-time display.
func (s *ClientService) CreateClient(ctx context.Context, req CreateClientRequest) (*models.Client, string, error) {
	clientID, err := utils.GenerateSecureToken(16)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate client_id: %w", err)
	}

	clientSecret, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate client_secret: %w", err)
	}

	hashedSecret, err := utils.HashPassword(clientSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash client_secret: %w", err)
	}

	client := &models.Client{
		ClientID:      clientID,
		ClientSecret:  hashedSecret,
		Name:          req.Name,
		RedirectURIs:  req.RedirectURIs,
		GrantTypes:    req.GrantTypes,
		ResponseTypes: req.ResponseTypes,
		Scopes:        req.Scopes,
		JWKSURL:       req.JWKSURL,
	}

	if err := s.clientStore.Create(ctx, client); err != nil {
		return nil, "", err
	}

	return client, clientSecret, nil
}

// ListClients retrieves all clients.
func (s *ClientService) ListClients(ctx context.Context) ([]models.Client, error) {
	return s.clientStore.List(ctx)
}

// DeleteClient deletes a client by its ID.
func (s *ClientService) DeleteClient(ctx context.Context, clientID string) error {
	return s.clientStore.Delete(ctx, clientID)
}

// GetClientByID retrieves a single client by its ID.
func (s *ClientService) GetClientByID(ctx context.Context, clientID string) (*models.Client, error) {
	return s.clientStore.GetByClientID(ctx, clientID)
}

// UpdateClient handles the business logic for updating an existing client.
func (s *ClientService) UpdateClient(ctx context.Context, clientID string, req UpdateClientRequest) (*models.Client, error) {
	// Fetch the existing client to ensure it exists.
	existingClient, err := s.clientStore.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, err // Will be ErrNotFound if it doesn't exist
	}

	// Update the fields from the request.
	existingClient.Name = req.Name
	existingClient.RedirectURIs = req.RedirectURIs
	existingClient.GrantTypes = req.GrantTypes
	existingClient.ResponseTypes = req.ResponseTypes
	existingClient.Scopes = req.Scopes
	existingClient.JWKSURL = req.JWKSURL

	// Persist the changes.
	if err := s.clientStore.Update(ctx, existingClient); err != nil {
		return nil, err
	}

	return existingClient, nil
}
