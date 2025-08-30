package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// IntrospectionHandler handles token introspection requests.
type IntrospectionHandler struct {
	logger        *slog.Logger
	clientService *services.ClientService
	jwtManager    *utils.JWTManager
}

// NewIntrospectionHandler creates a new IntrospectionHandler.
func NewIntrospectionHandler(logger *slog.Logger, clientService *services.ClientService, jwtManager *utils.JWTManager) *IntrospectionHandler {
	return &IntrospectionHandler{
		logger:        logger,
		clientService: clientService,
		jwtManager:    jwtManager,
	}
}

// Introspect is the main handler for the introspection endpoint.
func (h *IntrospectionHandler) Introspect(w http.ResponseWriter, r *http.Request) {
	// The introspection endpoint itself must be protected.
	// A resource server authenticates with its own client credentials.
	// We use HTTP Basic Authentication for this.
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		h.writeInactiveResponse(w)
		return
	}

	_, err := h.clientService.ValidateClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		h.writeInactiveResponse(w)
		return
	}

	// Now that the resource server is authenticated, we can inspect the token it sent.
	if err := r.ParseForm(); err != nil {
		h.writeInactiveResponse(w)
		return
	}
	tokenToInspect := r.PostForm.Get("token")
	if tokenToInspect == "" {
		h.writeInactiveResponse(w)
		return
	}

	// Verify the JWT.
	claims, err := h.jwtManager.VerifyToken(tokenToInspect)
	if err != nil {
		// If the token is invalid for any reason (bad signature, expired), it's inactive.
		h.writeInactiveResponse(w)
		return
	}

	// If we reach here, the token is valid. Respond with its claims.
	response := map[string]any{
		"active":     true,
		"scope":      strings.Join(claims.Scope, " "),
		"client_id":  claims.ClientID,
		"sub":        claims.Subject, // Subject (user_id or client_id)
		"exp":        claims.ExpiresAt.Unix(),
		"iat":        claims.IssuedAt.Unix(),
		"nbf":        claims.NotBefore.Unix(),
		"iss":        claims.Issuer,
		"aud":        claims.Audience,
		"jti":        claims.ID,
		"token_type": "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeInactiveResponse is a helper to return the standard response for an invalid token.
func (h *IntrospectionHandler) writeInactiveResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // The endpoint itself worked, so we return 200 OK.
	json.NewEncoder(w).Encode(map[string]bool{"active": false})
}
