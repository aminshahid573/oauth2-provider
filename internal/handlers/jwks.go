package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// JWKSHandler serves the JSON Web Key Set.
type JWKSHandler struct {
	logger     *slog.Logger
	jwtManager *utils.JWTManager
}

// NewJWKSHandler creates a new JWKSHandler.
func NewJWKSHandler(logger *slog.Logger, jwtManager *utils.JWTManager) *JWKSHandler {
	return &JWKSHandler{
		logger:     logger,
		jwtManager: jwtManager,
	}
}

// ServeJWKS is the HTTP handler for the JWKS endpoint.
func (h *JWKSHandler) ServeJWKS(w http.ResponseWriter, r *http.Request) {
	keySet, err := h.jwtManager.GetPublicKeySet()
	if err != nil {
		h.logger.Error("failed to get public key set", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// It's good practice to allow caching of the JWKS for a reasonable period.
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(keySet)
}
