package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/services"
)

// DiscoveryHandler serves the OAuth2 discovery document.
type DiscoveryHandler struct {
	logger        *slog.Logger
	clientService *services.ClientService
}

// NewDiscoveryHandler creates a new DiscoveryHandler.
func NewDiscoveryHandler(logger *slog.Logger, clientService *services.ClientService) *DiscoveryHandler {
	return &DiscoveryHandler{
		logger:        logger,
		clientService: clientService,
	}
}

// ServeDiscoveryDocument is the HTTP handler for the discovery endpoint.
func (h *DiscoveryHandler) ServeDiscoveryDocument(w http.ResponseWriter, r *http.Request) {
	baseURL := h.clientService.GetBaseURL()

	// Construct the discovery document.
	discoveryDoc := map[string]any{
		// --- Endpoint URLs ---
		"issuer":                        baseURL,
		"authorization_endpoint":        baseURL + "/oauth2/authorize",
		"token_endpoint":                baseURL + "/oauth2/token",
		"jwks_uri":                      baseURL + "/.well-known/jwks.json",
		"revocation_endpoint":           baseURL + "/oauth2/revoke",
		"introspection_endpoint":        baseURL + "/oauth2/introspect",
		"device_authorization_endpoint": baseURL + "/oauth2/device_authorization",

		// --- Supported Features ---
		"grant_types_supported": []string{
			models.GrantTypeAuthorizationCode,
			models.GrantTypeClientCredentials,
			models.GrantTypeRefreshToken,
			models.GrantTypeDeviceCode,
			models.GrantTypeJWTBearer,
		},
		"response_types_supported": []string{
			"code",
		},
		"scopes_supported": []string{
			"openid",
			"profile",
			"email",
			"offline",
			"api:read",
			"api:write",
		},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
		},
		"revocation_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
		},
		"introspection_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
		},
		"code_challenge_methods_supported": []string{
			"S256", // We will implement PKCE next
		},
		"id_token_signing_alg_values_supported": []string{
			"RS256",
		},
		"subject_types_supported": []string{
			"public",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(discoveryDoc); err != nil {
		h.logger.Error("failed to write discovery document", "error", err)
	}
}
