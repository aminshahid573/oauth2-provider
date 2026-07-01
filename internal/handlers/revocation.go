package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/aminshahid573/authexa/internal/services"
	"github.com/aminshahid573/authexa/internal/storage"
	"github.com/aminshahid573/authexa/internal/utils"
)

// RevocationHandler handles token revocation requests.
type RevocationHandler struct {
	logger          *slog.Logger
	clientService   *services.ClientService
	tokenService    *services.TokenService
	jwtManager      *utils.JWTManager
	revocationStore storage.RevocationStore
}

// NewRevocationHandler creates a new RevocationHandler.
func NewRevocationHandler(
	logger *slog.Logger,
	clientService *services.ClientService,
	tokenService *services.TokenService,
	jwtManager *utils.JWTManager,
	revocationStore storage.RevocationStore,
) *RevocationHandler {
	return &RevocationHandler{
		logger:          logger,
		clientService:   clientService,
		tokenService:    tokenService,
		jwtManager:      jwtManager,
		revocationStore: revocationStore,
	}
}

// Revoke is the main handler for the revocation endpoint.
// It supports revoking both refresh tokens (by deleting from the database)
// and JWT access tokens (by adding the JTI to a Redis revocation list).
// Conforms to RFC 7009.
func (h *RevocationHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate the client.
	// The client can authenticate using basic auth or by including credentials in the body.
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

	// 3. Try to handle as a JWT access token first.
	// If the token parses as a valid JWT, revoke it by adding its JTI to the
	// revocation list with a TTL equal to the token's remaining lifetime.
	claims, jwtErr := h.jwtManager.VerifyToken(tokenToRevoke)
	if jwtErr == nil {
		// Valid JWT access token -- verify client ownership via the aud/client_id claim.
		if claims.ClientID != client.ClientID {
			h.logger.Warn("client attempted to revoke an access token not belonging to it",
				"requesting_client", client.ClientID, "token_client", claims.ClientID)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Calculate remaining TTL for the revocation entry.
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 && claims.ID != "" {
			if err := h.revocationStore.Revoke(r.Context(), claims.ID, remaining); err != nil {
				h.logger.Error("failed to revoke access token JTI", "error", err, "jti", claims.ID)
			} else {
				h.logger.Info("access token revoked via JTI", "jti", claims.ID, "client_id", client.ClientID)
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 4. Not a JWT -- try as a refresh token (opaque, stored in database).
	signature := h.tokenService.HashToken(tokenToRevoke)
	token, err := h.tokenService.GetTokenBySignature(r.Context(), signature)
	if err != nil {
		// Token not found, but we still return 200 OK as per RFC.
		w.WriteHeader(http.StatusOK)
		return
	}

	// 5. Security check: Ensure the client revoking the token is the one it was issued to.
	if token.ClientID != client.ClientID {
		h.logger.Warn("client attempted to revoke a token not belonging to it",
			"requesting_client", client.ClientID, "token_client", token.ClientID)
		w.WriteHeader(http.StatusOK) // Return 200 OK to prevent information leakage.
		return
	}

	// 6. Delete the token from the database.
	if err := h.tokenService.DeleteTokenBySignature(r.Context(), signature); err != nil {
		h.logger.Error("failed to delete token during revocation", "error", err)
	}

	h.logger.Info("token revoked successfully", "client_id", client.ClientID)
	w.WriteHeader(http.StatusOK)
}
