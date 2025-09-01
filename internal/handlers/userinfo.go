package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserInfoHandler handles requests for user information.
type UserInfoHandler struct {
	logger     *slog.Logger
	jwtManager *utils.JWTManager
	userStore  storage.UserStore
}

// NewUserInfoHandler creates a new UserInfoHandler.
func NewUserInfoHandler(logger *slog.Logger, jwtManager *utils.JWTManager, userStore storage.UserStore) *UserInfoHandler {
	return &UserInfoHandler{
		logger:     logger,
		jwtManager: jwtManager,
		userStore:  userStore,
	}
}

// GetUserInfo is the handler for the userinfo endpoint.
func (h *UserInfoHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the token from the Authorization header.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeError(w, http.StatusUnauthorized, "invalid_token", "Authorization header missing")
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		h.writeError(w, http.StatusUnauthorized, "invalid_token", "Authorization header format must be Bearer {token}")
		return
	}
	tokenStr := parts[1]

	// 2. Validate the access token.
	claims, err := h.jwtManager.VerifyToken(tokenStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "invalid_token", "Access token is invalid or expired")
		return
	}

	// 3. Fetch the user from the database using the 'sub' (subject) claim.
	userID, err := bson.ObjectIDFromHex(claims.Subject)
	if err != nil {
		h.logger.Error("invalid subject claim in token", "subject", claims.Subject, "error", err)
		h.writeError(w, http.StatusUnauthorized, "invalid_token", "Invalid subject in token")
		return
	}
	user, err := h.userStore.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("user not found for valid token", "subject", claims.Subject, "error", err)
		h.writeError(w, http.StatusUnauthorized, "invalid_token", "User not found")
		return
	}

	// 4. Construct the response based on the scopes granted to the token.
	userInfo := make(map[string]any)
	// The 'sub' claim is always required.
	userInfo["sub"] = user.ID.Hex()

	scopeMap := make(map[string]bool)
	for _, s := range claims.Scope {
		scopeMap[s] = true
	}

	if scopeMap["profile"] {
		userInfo["name"] = user.Username // In a real app, you'd have more profile fields.
		userInfo["preferred_username"] = user.Username
	}
	if scopeMap["email"] {
		// Our user model doesn't have an email field, so we'll use the username as a placeholder.
		userInfo["email"] = user.Username + "@example.com"
		userInfo["email_verified"] = false // In a real app, this would be a real field.
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userInfo)
}

// writeError is a helper to send a standard OAuth2 error response for this endpoint.
func (h *UserInfoHandler) writeError(w http.ResponseWriter, status int, err, description string) {
	w.Header().Set("WWW-Authenticate", `Bearer error="`+err+`", error_description="`+description+`"`)
	http.Error(w, description, status)
}
