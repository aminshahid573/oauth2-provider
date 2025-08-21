package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/middleware"
	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// AuthHandler handles OAuth2 authorization and token requests.
type AuthHandler struct {
	logger        *slog.Logger
	templateCache utils.TemplateCache
	clientService *services.ClientService
	scopeService  *services.ScopeService
	tokenService  *services.TokenService // Add TokenService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	logger *slog.Logger,
	templateCache utils.TemplateCache,
	clientService *services.ClientService,
	scopeService *services.ScopeService,
	tokenService *services.TokenService, // Add TokenService
) *AuthHandler {
	return &AuthHandler{
		logger:        logger,
		templateCache: templateCache,
		clientService: clientService,
		scopeService:  scopeService,
		tokenService:  tokenService,
	}
}

// AuthorizeFlow is a single handler that routes to GET or POST logic.
func (h *AuthHandler) AuthorizeFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.showConsentPage(w, r)
	} else if r.Method == http.MethodPost {
		h.handleConsent(w, r)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// showConsentPage handles the GET request to the authorization endpoint.
func (h *AuthHandler) showConsentPage(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientID := queryParams.Get("client_id")
	redirectURI := queryParams.Get("redirect_uri")
	responseType := queryParams.Get("response_type")
	scope := queryParams.Get("scope")

	if clientID == "" || redirectURI == "" || responseType != "code" {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	client, err := h.clientService.GetClient(r.Context(), clientID)
	if err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrInvalidClient)
		return
	}

	// TODO: Validate redirect_uri against the client's registered URIs.

	requestedScopes := strings.Fields(scope)
	if !h.scopeService.ValidateScopes(requestedScopes) {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	scopeDetails := h.scopeService.GetScopeDetails(requestedScopes)
	data := map[string]any{
		"ClientName":  client.Name,
		"Scopes":      scopeDetails,
		"QueryParams": queryParams,
	}
	h.templateCache.Render(w, r, "consent.html", data)
}

// handleConsent handles the POST request from the consent form.
func (h *AuthHandler) handleConsent(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		utils.HandleError(w, r, h.logger, utils.ErrInternal) // Should not happen if middleware is applied
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	// Re-validate parameters from the form
	clientID := r.PostForm.Get("client_id")
	redirectURIStr := r.PostForm.Get("redirect_uri")
	state := r.PostForm.Get("state")
	scope := r.PostForm.Get("scope")

	redirectURL, err := url.Parse(redirectURIStr)
	if err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}
	query := redirectURL.Query()

	// If the user denied access, redirect with an error.
	if r.PostForm.Get("consent") != "allow" {
		query.Set("error", "access_denied")
		if state != "" {
			query.Set("state", state)
		}
		redirectURL.RawQuery = query.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	// If the user allowed access, generate an authorization code.
	requestedScopes := strings.Fields(scope)
	code, err := h.tokenService.GenerateAndStoreAuthorizationCode(r.Context(), user.ID.Hex(), clientID, requestedScopes)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	query.Set("code", code)
	if state != "" {
		query.Set("state", state)
	}
	redirectURL.RawQuery = query.Encode()
	http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
}

// Token handles POST requests to the token endpoint.
func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("failed to parse token request form", "error", err)
		h.writeTokenError(w, "invalid_request", "The request is malformed.")
		return
	}

	grantType := r.PostForm.Get("grant_type")
	h.logger.Info("token request received", "grant_type", grantType) // Added for debugging

	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(w, r)
	case "client_credentials":
		h.handleClientCredentialsGrant(w, r)
	case "refresh_token":
		h.handleRefreshTokenGrant(w, r)
	default:
		h.logger.Warn("unsupported grant type requested", "grant_type", grantType)
		h.writeTokenError(w, "unsupported_grant_type", "The authorization grant type is not supported.")
	}
}

// handleRefreshTokenGrant processes the refresh_token grant type.
func (h *AuthHandler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr := r.PostForm.Get("refresh_token")
	clientID := r.PostForm.Get("client_id")
	clientSecret := r.PostForm.Get("client_secret")

	// 1. Authenticate the client.
	client, err := h.clientService.ValidateClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		h.writeTokenError(w, "invalid_client", "Client authentication failed.")
		return
	}

	// 2. Validate the refresh token.
	refreshToken, err := h.tokenService.ValidateAndConsumeRefreshToken(r.Context(), refreshTokenStr)
	if err != nil {
		h.writeTokenError(w, "invalid_grant", "The refresh token is invalid or expired.")
		return
	}

	// 3. Ensure the token was issued to this client.
	if refreshToken.ClientID != client.ClientID {
		h.writeTokenError(w, "invalid_grant", "The refresh token was not issued to this client.")
		return
	}

	// 4. Generate a new access token.
	// The new access token can have the same or a subset of the original scopes.
	// For simplicity, we'll grant the same scopes.
	accessToken, err := h.tokenService.GenerateAccessToken(refreshToken.UserID, client.ClientID, refreshToken.Scopes)
	if err != nil {
		h.logger.Error("failed to generate access token from refresh token", "error", err)
		h.writeTokenError(w, "server_error", "The server encountered an error.")
		return
	}

	// 5. Respond with the new token.
	// Note: A new refresh token is NOT issued in this simple implementation.
	// For refresh token rotation, you would generate and include a new refresh token here.
	tokenResponse := map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        strings.Join(refreshToken.Scopes, " "),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

// handleAuthorizationCodeGrant processes the authorization_code grant type.
func (h *AuthHandler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.PostForm.Get("code")
	// redirectURI := r.PostForm.Get("redirect_uri") // Commented out as per previous step
	clientID := r.PostForm.Get("client_id")
	clientSecret := r.PostForm.Get("client_secret")

	client, err := h.clientService.ValidateClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		h.writeTokenError(w, "invalid_client", "Client authentication failed.")
		return
	}

	authCodeToken, err := h.tokenService.ValidateAndConsumeAuthCode(r.Context(), code)
	if err != nil {
		h.writeTokenError(w, "invalid_grant", "The authorization code is invalid or expired.")
		return
	}

	if authCodeToken.ClientID != client.ClientID {
		h.writeTokenError(w, "invalid_grant", "The authorization code was not issued to this client.")
		return
	}

	accessToken, err := h.tokenService.GenerateAccessToken(authCodeToken.UserID, client.ClientID, authCodeToken.Scopes)
	if err != nil {
		h.logger.Error("failed to generate access token", "error", err)
		h.writeTokenError(w, "server_error", "The server encountered an error.")
		return
	}

	refreshToken, err := h.tokenService.GenerateAndStoreRefreshToken(r.Context(), authCodeToken.UserID, client.ClientID, authCodeToken.Scopes)
	if err != nil {
		h.logger.Error("failed to generate refresh token", "error", err)
		h.writeTokenError(w, "server_error", "The server encountered an error.")
		return
	}

	tokenResponse := map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"scope":         strings.Join(authCodeToken.Scopes, " "),
		"refresh_token": refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

// handleClientCredentialsGrant processes the client_credentials grant type.
func (h *AuthHandler) handleClientCredentialsGrant(w http.ResponseWriter, r *http.Request) {
	clientID := r.PostForm.Get("client_id")
	clientSecret := r.PostForm.Get("client_secret")
	scope := r.PostForm.Get("scope")

	client, err := h.clientService.ValidateClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		h.writeTokenError(w, "invalid_client", "Client authentication failed.")
		return
	}

	if !slices.Contains(client.GrantTypes, models.GrantTypeClientCredentials) {
		h.writeTokenError(w, "unauthorized_client", "The client is not authorized to use this grant type.")
		return
	}

	requestedScopes := strings.Fields(scope)
	if len(requestedScopes) == 0 {
		requestedScopes = client.Scopes
	} else {
		for _, s := range requestedScopes {
			if !slices.Contains(client.Scopes, s) {
				h.writeTokenError(w, "invalid_scope", "The requested scope is invalid, unknown, or malformed.")
				return
			}
		}
	}

	accessToken, err := h.tokenService.GenerateAccessToken(client.ClientID, client.ClientID, requestedScopes)
	if err != nil {
		h.logger.Error("failed to generate access token for client credentials", "error", err)
		h.writeTokenError(w, "server_error", "The server encountered an error.")
		return
	}

	tokenResponse := map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        strings.Join(requestedScopes, " "),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

// writeTokenError is a helper to send a standard OAuth2 error response.
func (h *AuthHandler) writeTokenError(w http.ResponseWriter, err, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest) // Most token errors are 400
	json.NewEncoder(w).Encode(map[string]string{
		"error":             err,
		"error_description": description,
	})
}
