package handlers

import (
	"log/slog"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/services"
)

// RevocationHandler handles token revocation requests.
type RevocationHandler struct {
	logger        *slog.Logger
	clientService *services.ClientService
	tokenService  *services.TokenService
}

// NewRevocationHandler creates a new RevocationHandler.
func NewRevocationHandler(logger *slog.Logger, clientService *services.ClientService, tokenService *services.TokenService) *RevocationHandler {
	return &RevocationHandler{
		logger:        logger,
		clientService: clientService,
		tokenService:  tokenService,
	}
}

// Revoke is the main handler for the revocation endpoint.
func (h *RevocationHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate the client.
	// The client can authenticate using basic auth or by including credentials in the body.
	// We'll support basic auth for simplicity.
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		// Fallback to checking form parameters
		clientID = r.PostFormValue("client_id")
		clientSecret = r.PostFormValue("client_secret")
	}

	client, err := h.clientService.ValidateClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		// RFC 7009 says to return 200 OK even for invalid clients to prevent snooping.
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. Get the token to revoke from the form body.
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokenToRevoke := r.PostForm.Get("token")
	if tokenToRevoke == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 3. For now, we only support revoking refresh tokens.
	// Access tokens are stateless JWTs and will expire on their own.
	// Revoking the refresh token prevents new access tokens from being issued.
	signature := h.tokenService.HashToken(tokenToRevoke)
	token, err := h.tokenService.GetTokenBySignature(r.Context(), signature)
	if err != nil {
		// Token not found, but we still return 200 OK as per RFC.
		w.WriteHeader(http.StatusOK)
		return
	}

	// 4. Security check: Ensure the client revoking the token is the one it was issued to.
	if token.ClientID != client.ClientID {
		h.logger.Warn("client attempted to revoke a token not belonging to it", "requesting_client", client.ClientID, "token_client", token.ClientID)
		w.WriteHeader(http.StatusOK) // Return 200 OK to prevent information leakage.
		return
	}

	// 5. Delete the token from the database.
	err = h.tokenService.DeleteTokenBySignature(r.Context(), signature)
	if err != nil {
		h.logger.Error("failed to delete token during revocation", "error", err)
		// Even if deletion fails, we should probably not signal an error to the client.
	}

	h.logger.Info("token revoked successfully", "client_id", client.ClientID)
	w.WriteHeader(http.StatusOK)
}
