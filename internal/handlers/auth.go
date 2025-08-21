package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/middleware"
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
