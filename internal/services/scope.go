package services

import "github.com/aminshahid573/oauth2-provider/internal/models"

// ScopeService provides logic for managing OAuth2 scopes.
type ScopeService struct {
	// For now, we use a simple map. This could be backed by a database.
	availableScopes map[string]string
}

// NewScopeService creates a new ScopeService.
func NewScopeService() *ScopeService {
	return &ScopeService{
		availableScopes: map[string]string{
			"openid":  "Access your user identifier.",
			"profile": "Read your basic profile information.",
			"email":   "Access your email address.",
			"offline": "Allow the application to refresh tokens.",
		},
	}
}

// GetScopeDetails retrieves the details for a list of scope names.
func (s *ScopeService) GetScopeDetails(scopeNames []string) []models.Scope {
	var details []models.Scope
	for _, name := range scopeNames {
		if description, ok := s.availableScopes[name]; ok {
			details = append(details, models.Scope{
				Name:        name,
				Description: description,
			})
		}
	}
	return details
}

// ValidateScopes checks if all requested scopes are valid.
func (s *ScopeService) ValidateScopes(requestedScopes []string) bool {
	for _, rs := range requestedScopes {
		if _, ok := s.availableScopes[rs]; !ok {
			return false // Found an invalid scope
		}
	}
	return true
}
