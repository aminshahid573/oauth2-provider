// File: internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/middleware"
	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// AuthHandler handles OAuth2 authorization and token requests.
type AuthHandler struct {
	logger        *slog.Logger
	templateCache utils.TemplateCache
	clientService *services.ClientService
	scopeService  *services.ScopeService
	tokenService  *services.TokenService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	logger *slog.Logger,
	templateCache utils.TemplateCache,
	clientService *services.ClientService,
	scopeService *services.ScopeService,
	tokenService *services.TokenService,
) *AuthHandler {
	return &AuthHandler{
		logger:        logger,
		templateCache: templateCache,
		clientService: clientService,
		scopeService:  scopeService,
		tokenService:  tokenService,
	}
}

// --- Standard Authorization Code Flow ---

// AuthorizeFlow is a single handler that routes to GET or POST logic for the standard user-facing flow.
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

	codeChallenge := queryParams.Get("code_challenge")
	codeChallengeMethod := queryParams.Get("code_challenge_method")

	if codeChallenge != "" {
		if codeChallengeMethod != "S256" {
			utils.HandleError(w, r, h.logger, h.templateCache, &utils.AppError{Code: "invalid_request", Message: "code_challenge_method must be S256.", HTTPStatus: http.StatusBadRequest})
			return
		}
	}

	if clientID == "" || redirectURI == "" || responseType != "code" {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrBadRequest)
		return
	}

	client, err := h.clientService.GetClient(r.Context(), clientID)
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrInvalidClient)
		return
	}

	// The provided redirect_uri MUST be one of the URIs registered by the client.
	if !slices.Contains(client.RedirectURIs, redirectURI) {
		h.logger.Warn("invalid redirect_uri provided", "client_id", clientID, "provided_uri", redirectURI)
		utils.HandleError(w, r, h.logger, h.templateCache, &utils.AppError{Code: "invalid_request", Message: "The provided redirect_uri is not registered for this client.", HTTPStatus: http.StatusBadRequest})
		return
	}

	requestedScopes := strings.Fields(scope)
	if !h.scopeService.ValidateScopes(requestedScopes) {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrBadRequest)
		return
	}

	scopeDetails := h.scopeService.GetScopeDetails(requestedScopes)
	data := map[string]any{
		"ClientName":  client.Name,
		"Scopes":      scopeDetails,
		"QueryParams": queryParams,
	}
	h.templateCache.Render(w, r, "base.html", "consent.html", data)
}

// handleConsent handles the POST request from the consent form.
func (h *AuthHandler) handleConsent(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrInternal)
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrBadRequest)
		return
	}

	clientID := r.PostForm.Get("client_id")
	redirectURIStr := r.PostForm.Get("redirect_uri")
	state := r.PostForm.Get("state")
	scope := r.PostForm.Get("scope")

	codeChallenge := r.PostForm.Get("code_challenge")
	codeChallengeMethod := r.PostForm.Get("code_challenge_method")

	redirectURL, err := url.Parse(redirectURIStr)
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrBadRequest)
		return
	}
	query := redirectURL.Query()

	if r.PostForm.Get("consent") != "allow" {
		query.Set("error", "access_denied")
		if state != "" {
			query.Set("state", state)
		}
		redirectURL.RawQuery = query.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	requestedScopes := strings.Fields(scope)
	code, err := h.tokenService.GenerateAndStoreAuthorizationCode(r.Context(), user.ID.Hex(), clientID, requestedScopes)
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, err)
		return
	}

	if codeChallenge != "" && codeChallengeMethod == "S256" {
		err := h.tokenService.StorePKCEChallenge(r.Context(), code, codeChallenge)
		if err != nil {
			utils.HandleError(w, r, h.logger, h.templateCache, err)
			return
		}
	}

	query.Set("code", code)
	if state != "" {
		query.Set("state", state)
	}
	redirectURL.RawQuery = query.Encode()
	http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
}

// --- Device Authorization Flow ---

// DeviceAuthorization handles POST requests to the device_authorization endpoint.
func (h *AuthHandler) DeviceAuthorization(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrBadRequest)
		return
	}

	clientID := r.PostForm.Get("client_id")
	scope := r.PostForm.Get("scope")

	client, err := h.clientService.GetClient(r.Context(), clientID)
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrInvalidClient)
		return
	}

	deviceCode, userCode, err := h.tokenService.GenerateAndStoreDeviceCode(r.Context(), client.ClientID, strings.Fields(scope))
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, err)
		return
	}

	verificationURI := h.clientService.GetBaseURL() + "/device"
	response := map[string]any{
		"device_code":      deviceCode,
		"user_code":        userCode,
		"verification_uri": verificationURI,
		"expires_in":       int(services.DeviceCodeLifespan.Seconds()),
		"interval":         5,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeviceConsentFlow handles the user-facing part of the device flow after code entry.
func (h *AuthHandler) DeviceConsentFlow(w http.ResponseWriter, r *http.Request) {
	userCode := r.URL.Query().Get("user_code")
	if userCode == "" {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrBadRequest)
		return
	}

	// Find the device code to get client and scope info
	token, err := h.tokenService.GetTokenByUserCode(r.Context(), userCode)
	if err != nil {
		// TODO: Render a proper error page
		utils.HandleError(w, r, h.logger, h.templateCache, &utils.AppError{Code: "INVALID_CODE", Message: "Invalid or expired code.", HTTPStatus: http.StatusBadRequest})
		return
	}

	client, err := h.clientService.GetClient(r.Context(), token.ClientID)
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrInvalidClient)
		return
	}

	scopeDetails := h.scopeService.GetScopeDetails(token.Scopes)
	data := map[string]any{
		"ClientName": client.Name,
		"Scopes":     scopeDetails,
		"UserCode":   userCode, // Pass the user_code to the form
	}
	h.templateCache.Render(w, r, "admin.html", "consent_device.html", data)
}

// HandleDeviceConsent handles the POST from the device consent page.
func (h *AuthHandler) HandleDeviceConsent(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrInternal)
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, utils.ErrBadRequest)
		return
	}

	userCode := r.PostForm.Get("user_code")
	if r.PostForm.Get("consent") != "allow" {
		// TODO: Show a "denied" page
		w.Write([]byte("Authorization denied."))
		return
	}

	_, err := h.tokenService.ApproveDeviceCode(r.Context(), userCode, user.ID.Hex())
	if err != nil {
		// TODO: Show an error page
		utils.HandleError(w, r, h.logger, h.templateCache, err)
		return
	}

	// Show a success page to the user
	h.templateCache.Render(w, r, "admin.html", "device_success.html", nil)
}

// --- Token Endpoint ---

// Token handles POST requests to the token endpoint for all grant types.
func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("failed to parse token request form", "error", err)
		h.writeTokenError(w, "invalid_request", "The request is malformed.")
		return
	}

	grantType := r.PostForm.Get("grant_type")
	h.logger.Info("token request received", "grant_type", grantType)

	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(w, r)
	case "client_credentials":
		h.handleClientCredentialsGrant(w, r)
	case "refresh_token":
		h.handleRefreshTokenGrant(w, r)
	case "urn:ietf:params:oauth:grant-type:device_code":
		h.handleDeviceCodeGrant(w, r)
	case "urn:ietf:params:oauth:grant-type:jwt-bearer":
		h.handleJWTBearerGrant(w, r)
	default:
		h.logger.Warn("unsupported grant type requested", "grant_type", grantType)
		h.writeTokenError(w, "unsupported_grant_type", "The authorization grant type is not supported.")
	}
}

// handleJWTBearerGrant processes the jwt-bearer grant type.
func (h *AuthHandler) handleJWTBearerGrant(w http.ResponseWriter, r *http.Request) {
	assertionStr := r.PostForm.Get("assertion")
	if assertionStr == "" {
		h.writeTokenError(w, "invalid_request", "Missing assertion parameter.")
		return
	}

	// 1. Parse the assertion without validating the signature yet.
	// This allows us to read the claims and find out which client is making the request.
	token, _, err := new(jwt.Parser).ParseUnverified(assertionStr, jwt.MapClaims{})
	if err != nil {
		h.writeTokenError(w, "invalid_grant", "Assertion is malformed.")
		return
	}

	// 2. Extract client_id from the 'iss' (issuer) claim.
	clientID, err := token.Claims.GetIssuer()
	if err != nil {
		h.writeTokenError(w, "invalid_grant", "Assertion is missing 'iss' claim.")
		return
	}

	// 3. Fetch the client from our database.
	client, err := h.clientService.GetClient(r.Context(), clientID)
	if err != nil {
		h.writeTokenError(w, "invalid_client", "Client not found.")
		return
	}

	// 4. Validate that the client is allowed to use this grant type.
	if !slices.Contains(client.GrantTypes, models.GrantTypeJWTBearer) {
		h.writeTokenError(w, "unauthorized_client", "The client is not authorized to use this grant type.")
		return
	}
	if client.JWKSURL == "" {
		h.writeTokenError(w, "unauthorized_client", "Client is not configured for JWT Bearer grant (missing JWKS URL).")
		return
	}

	// 5. Fetch the client's public key from their JWKS URL and validate the assertion signature.
	keySet, err := jwk.Fetch(r.Context(), client.JWKSURL)
	if err != nil {
		h.logger.Error("failed to fetch client JWKS", "url", client.JWKSURL, "error", err)
		h.writeTokenError(w, "invalid_grant", "Could not fetch client's public keys.")
		return
	}

	// The jwt.Parse function will find the correct key from the key set based on the 'kid' header in the JWT.
	parsedToken, err := jwt.Parse(assertionStr, func(t *jwt.Token) (interface{}, error) {
		// Find the key that matches the 'kid' in the JWT header.
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		key, found := keySet.LookupKeyID(kid)
		if !found {
			return nil, jwt.ErrTokenUnverifiable
		}
		var pubKey interface{}
		if err := key.Raw(&pubKey); err != nil {
			return nil, jwt.ErrTokenUnverifiable
		}
		return pubKey, nil
	})

	if err != nil || !parsedToken.Valid {
		h.writeTokenError(w, "invalid_grant", "Assertion validation failed.")
		return
	}

	// 6. Validate other claims ('aud', 'exp').
	// The 'aud' (audience) claim MUST be our token endpoint URL.
	expectedAudience := h.clientService.GetBaseURL() + "/oauth2/token"
	audience, err := parsedToken.Claims.GetAudience()
	if err != nil || len(audience) == 0 || audience[0] != expectedAudience {
		h.writeTokenError(w, "invalid_grant", "Invalid 'aud' claim in assertion.")
		return
	}

	// 7. If everything is valid, issue an access token.
	requestedScopes := client.Scopes // For this flow, grant all allowed scopes.
	accessToken, err := h.tokenService.GenerateAccessToken(client.ClientID, client.ClientID, requestedScopes)
	if err != nil {
		h.logger.Error("failed to generate access token for JWT bearer", "error", err)
		h.writeTokenError(w, "server_error", "The server encountered an error.")
		return
	}

	tokenResponse := map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(h.tokenService.GetAccessTokenLifespan().Seconds()),
		"scope":        strings.Join(requestedScopes, " "),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

// handleAuthorizationCodeGrant processes the authorization_code grant type.
func (h *AuthHandler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.PostForm.Get("code")
	redirectURI := r.PostForm.Get("redirect_uri")
	clientID := r.PostForm.Get("client_id")
	clientSecret := r.PostForm.Get("client_secret")
	codeVerifier := r.PostForm.Get("code_verifier")

	client, err := h.clientService.ValidateClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		h.writeTokenError(w, "invalid_client", "Client authentication failed.")
		return
	}

	// The redirect_uri in the token request MUST match the one from the authorization request.
	// While we don't store it with the code, we can validate it's a valid one for the client.
	if !slices.Contains(client.RedirectURIs, redirectURI) {
		h.writeTokenError(w, "invalid_grant", "The redirect_uri does not match the one registered for the client.")
		return
	}

	err = h.tokenService.ValidatePKCE(r.Context(), code, codeVerifier)
	if err != nil {
		// err could be ErrNotFound (no challenge stored) or a mismatch.
		h.writeTokenError(w, "invalid_grant", err.Error())
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
		"expires_in":    int(h.tokenService.GetAccessTokenLifespan().Seconds()),
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
		"expires_in":   int(h.tokenService.GetAccessTokenLifespan().Seconds()),
		"scope":        strings.Join(requestedScopes, " "),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

// handleRefreshTokenGrant processes the refresh_token grant type.
func (h *AuthHandler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr := r.PostForm.Get("refresh_token")
	clientID := r.PostForm.Get("client_id")
	clientSecret := r.PostForm.Get("client_secret")

	client, err := h.clientService.ValidateClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		h.writeTokenError(w, "invalid_client", "Client authentication failed.")
		return
	}

	refreshToken, err := h.tokenService.ValidateAndConsumeRefreshToken(r.Context(), refreshTokenStr)
	if err != nil {
		h.writeTokenError(w, "invalid_grant", "The refresh token is invalid or expired.")
		return
	}

	if refreshToken.ClientID != client.ClientID {
		h.writeTokenError(w, "invalid_grant", "The refresh token was not issued to this client.")
		return
	}

	accessToken, err := h.tokenService.GenerateAccessToken(refreshToken.UserID, client.ClientID, refreshToken.Scopes)
	if err != nil {
		h.logger.Error("failed to generate access token from refresh token", "error", err)
		h.writeTokenError(w, "server_error", "The server encountered an error.")
		return
	}

	tokenResponse := map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(h.tokenService.GetAccessTokenLifespan().Seconds()),
		"scope":        strings.Join(refreshToken.Scopes, " "),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

// handleDeviceCodeGrant processes the device_code grant type.
func (h *AuthHandler) handleDeviceCodeGrant(w http.ResponseWriter, r *http.Request) {
	deviceCode := r.PostForm.Get("device_code")
	clientID := r.PostForm.Get("client_id")

	signature := h.tokenService.HashToken(deviceCode)
	token, err := h.tokenService.GetTokenBySignature(r.Context(), signature)
	if err != nil || token.ClientID != clientID {
		h.writeTokenError(w, "invalid_grant", "Device code is invalid, expired, or not for this client.")
		return
	}

	if time.Now().After(token.ExpiresAt) {
		h.writeTokenError(w, "expired_token", "The device code has expired.")
		return
	}

	if !token.Approved {
		h.writeTokenError(w, "authorization_pending", "User has not yet approved the request.")
		return
	}

	_ = h.tokenService.DeleteTokenBySignature(r.Context(), signature)

	accessToken, err := h.tokenService.GenerateAccessToken(token.UserID, token.ClientID, token.Scopes)
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, err)
		return
	}

	refreshToken, err := h.tokenService.GenerateAndStoreRefreshToken(r.Context(), token.UserID, token.ClientID, token.Scopes)
	if err != nil {
		utils.HandleError(w, r, h.logger, h.templateCache, err)
		return
	}

	tokenResponse := map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(h.tokenService.GetAccessTokenLifespan().Seconds()),
		"scope":         strings.Join(token.Scopes, " "),
		"refresh_token": refreshToken,
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
	status := http.StatusBadRequest
	if err == "authorization_pending" {
		// This is a special case for device flow polling.
		status = http.StatusPreconditionRequired // Using 428 as a distinct status
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             err,
		"error_description": description,
	})
}
