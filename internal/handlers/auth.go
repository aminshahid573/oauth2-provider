package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// AuthHandler handles OAuth2 authorization and token requests.
type AuthHandler struct {
	logger         *slog.Logger
	templateCache  utils.TemplateCache
	clientService  *services.ClientService
	scopeService   *services.ScopeService
	sessionService *services.SessionService
	// We will add more services later
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	logger *slog.Logger,
	templateCache utils.TemplateCache,
	clientService *services.ClientService,
	scopeService *services.ScopeService,
	sessionService *services.SessionService,
) *AuthHandler {
	return &AuthHandler{
		logger:         logger,
		templateCache:  templateCache,
		clientService:  clientService,
		scopeService:   scopeService,
		sessionService: sessionService,
	}
}

// Authorize handles the GET request to the authorization endpoint.
// It validates the request and shows the login or consent page.
func (h *AuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	// TODO: Check if user is logged in via session cookie.
	// If not logged in, redirect to /login with the original query params.

	// 1. Extract and validate query parameters.
	queryParams := r.URL.Query()
	clientID := queryParams.Get("client_id")
	redirectURI := queryParams.Get("redirect_uri")
	responseType := queryParams.Get("response_type")
	scope := queryParams.Get("scope")
	// state := queryParams.Get("state") // Important for security

	// Basic validation
	if clientID == "" || redirectURI == "" || responseType != "code" {
		// In a real scenario, you would redirect to the client's redirect_uri with an error.
		// For now, we show a simple error.
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	// 2. Validate the client.
	client, err := h.clientService.GetClient(r.Context(), clientID)
	if err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrInvalidClient)
		return
	}

	// TODO: Validate redirect_uri against the client's registered URIs.

	// 3. Validate scopes.
	requestedScopes := strings.Fields(scope)
	if !h.scopeService.ValidateScopes(requestedScopes) {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	// 4. Render the consent page.
	scopeDetails := h.scopeService.GetScopeDetails(requestedScopes)

	data := map[string]any{
		"ClientName":  client.Name,
		"Scopes":      scopeDetails,
		"QueryParams": queryParams, // Pass all params to the form
	}

	h.templateCache.Render(w, r, "consent.html", data)
}
