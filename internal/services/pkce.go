package services

import (
	"context"
	"crypto/subtle"
	"fmt"

	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// PKCEService provides logic for PKCE validation.
type PKCEService struct {
	pkceStore storage.PKCEStore
}

// NewPKCEService creates a new PKCEService.
func NewPKCEService(pkceStore storage.PKCEStore) *PKCEService {
	return &PKCEService{pkceStore: pkceStore}
}

// Validate checks if the provided code verifier matches the stored challenge.
func (s *PKCEService) Validate(ctx context.Context, code, verifier string) error {
	// Retrieve the stored challenge associated with the authorization code.
	storedChallenge, err := s.pkceStore.Get(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to retrieve PKCE challenge: %w", err)
	}

	// IMPORTANT: After retrieving, immediately delete the challenge to prevent reuse.
	defer s.pkceStore.Delete(ctx, code)

	// Calculate the challenge from the verifier provided by the client.
	generatedChallenge := utils.GeneratePKCEChallengeS256(verifier)

	// Compare the stored challenge with the generated one in constant time to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(storedChallenge), []byte(generatedChallenge)) != 1 {
		return utils.ErrBadRequest
	}

	return nil
}
