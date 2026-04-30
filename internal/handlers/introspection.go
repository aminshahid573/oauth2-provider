package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aminshahid573/authexa/internal/models"
	"github.com/aminshahid573/authexa/internal/services"
	"github.com/aminshahid573/authexa/internal/utils"
)

// IntrospectionHandler handles token introspection requests.
type IntrospectionHandler struct {
	logger        *slog.Logger
	clientService *services.ClientService
	tokenService  *services.TokenService
	jwtManager    *utils.JWTManager
}

// NewIntrospectionHandler creates a new IntrospectionHandler.
func NewIntrospectionHandler(logger *slog.Logger, clientService *services.ClientService, tokenService *services.TokenService, jwtManager *utils.JWTManager) *IntrospectionHandler {
	return &IntrospectionHandler{
		logger:        logger,
		clientService: clientService,
		tokenService:  tokenService,
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

	// First, try verifying it as a JWT Access Token.
	claims, err := h.jwtManager.VerifyToken(tokenToInspect)
	if err == nil {
		// Valid JWT Access Token
		response := map[string]any{
			"active":     true,
			"scope":      strings.Join(claims.Scope, " "),
			"client_id":  claims.ClientID,
			"sub":        claims.Subject,
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
		return
	}

	// If not a valid JWT, check if it's a Refresh Token in the database.
	signature := h.tokenService.HashToken(tokenToInspect)
	token, err := h.tokenService.GetTokenBySignature(r.Context(), signature)
	if err == nil && token.Type == models.TokenTypeRefreshToken {
		// Check if it's expired
		if time.Now().Before(token.ExpiresAt) {
			response := map[string]any{
				"active":     true,
				"client_id":  token.ClientID,
				"scope":      strings.Join(token.Scopes, " "),
				"sub":        token.UserID,
				"exp":        token.ExpiresAt.Unix(),
				"token_type": "Refresh Token",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// If we reach here, it's neither a valid access token nor a valid refresh token.
	h.writeInactiveResponse(w)
}

// writeInactiveResponse is a helper to return the standard response for an invalid token.
func (h *IntrospectionHandler) writeInactiveResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // The endpoint itself worked, so we return 200 OK.
	json.NewEncoder(w).Encode(map[string]bool{"active": false})
}
