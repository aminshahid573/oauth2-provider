package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/middleware"
	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"github.com/go-playground/validator/v10"
)

// AdminHandler handles requests for the admin API.
type AdminHandler struct {
	logger           *slog.Logger
	clientService    *services.ClientService
	userService      *services.UserService
	validate         *validator.Validate
	dashboardService *services.DashboardService
	auditService     *services.AuditService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(logger *slog.Logger, clientService *services.ClientService, userService *services.UserService, dashboardService *services.DashboardService, auditService *services.AuditService) *AdminHandler {
	return &AdminHandler{
		logger:           logger,
		clientService:    clientService,
		userService:      userService,
		validate:         validator.New(),
		dashboardService: dashboardService,
		auditService:     auditService,
	}
}

// GetStats handles the request to get dashboard statistics.
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.dashboardService.GetStats(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ListClients handles the request to list all clients.
func (h *AdminHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.clientService.ListClients(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	// For security, we don't expose the secret hash in the list view.
	type clientResponse struct {
		ClientID      string   `json:"client_id"`
		Name          string   `json:"name"`
		RedirectURIs  []string `json:"redirect_uris"`
		GrantTypes    []string `json:"grant_types"`
		ResponseTypes []string `json:"response_types"`
		Scopes        []string `json:"scopes"`
	}

	response := make([]clientResponse, len(clients))
	for i, c := range clients {
		response[i] = clientResponse{
			ClientID:      c.ClientID,
			Name:          c.Name,
			RedirectURIs:  c.RedirectURIs,
			GrantTypes:    c.GrantTypes,
			ResponseTypes: c.ResponseTypes,
			Scopes:        c.Scopes,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateClient handles the request to create a new client.
func (h *AdminHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req services.CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.HandleError(w, r, h.logger, &utils.AppError{Code: "VALIDATION_ERROR", Message: err.Error(), HTTPStatus: http.StatusBadRequest})
		return
	}

	client, plaintextSecret, err := h.clientService.CreateClient(r.Context(), req)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	response := map[string]any{
		"client_id":      client.ClientID,
		"client_secret":  plaintextSecret, // IMPORTANT: Show the secret only on creation
		"name":           client.Name,
		"redirect_uris":  client.RedirectURIs,
		"grant_types":    client.GrantTypes,
		"response_types": client.ResponseTypes,
		"scopes":         client.Scopes,
		"jwks_url":       client.JWKSURL,
	}

	user, _ := middleware.GetUserFromContext(r)
	eventData := services.RecordEventData{
		EventType: models.ClientCreated,
		ActorID:   user.ID.Hex(), // The admin who performed the action
		TargetID:  client.ClientID,
		IPAddress: middleware.GetClientIP(r),
		UserAgent: r.UserAgent(),
		Details:   "Admin created new client via API.",
	}
	_ = h.auditService.Record(r.Context(), eventData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// DeleteClient handles the request to delete a client.
func (h *AdminHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientID")
	if clientID == "" {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	if err := h.clientService.DeleteClient(r.Context(), clientID); err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetClient handles the request to retrieve a single client.
func (h *AdminHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientID")
	client, err := h.clientService.GetClientByID(r.Context(), clientID)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	// We don't expose the secret hash.
	response := map[string]any{
		"client_id":      client.ClientID,
		"name":           client.Name,
		"redirect_uris":  client.RedirectURIs,
		"grant_types":    client.GrantTypes,
		"response_types": client.ResponseTypes,
		"scopes":         client.Scopes,
		"jwks_url":       client.JWKSURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateClient handles the request to update a client.
func (h *AdminHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientID")

	var req services.UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.HandleError(w, r, h.logger, &utils.AppError{Code: "VALIDATION_ERROR", Message: err.Error(), HTTPStatus: http.StatusBadRequest})
		return
	}

	updatedClient, err := h.clientService.UpdateClient(r.Context(), clientID, req)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	// Respond with the updated client data (no secret).
	response := map[string]any{
		"client_id":      updatedClient.ClientID,
		"name":           updatedClient.Name,
		"redirect_uris":  updatedClient.RedirectURIs,
		"grant_types":    updatedClient.GrantTypes,
		"response_types": updatedClient.ResponseTypes,
		"scopes":         updatedClient.Scopes,
		"jwks_url":       updatedClient.JWKSURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListUsers handles the request to list all users.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	// For security, never expose the password hash.
	type userResponse struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	response := make([]userResponse, len(users))
	for i, u := range users {
		response[i] = userResponse{
			ID:       u.ID.Hex(),
			Username: u.Username,
			Role:     u.Role,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateUser handles the request to create a new user.
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req services.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.HandleError(w, r, h.logger, &utils.AppError{Code: "VALIDATION_ERROR", Message: err.Error(), HTTPStatus: http.StatusBadRequest})
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	response := map[string]any{
		"id":       user.ID.Hex(),
		"username": user.Username,
		"role":     user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUser handles the request to retrieve a single user.
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	response := map[string]any{
		"id":       user.ID.Hex(),
		"username": user.Username,
		"role":     user.Role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUser handles the request to update a user.
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	var req services.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.HandleError(w, r, h.logger, &utils.AppError{Code: "VALIDATION_ERROR", Message: err.Error(), HTTPStatus: http.StatusBadRequest})
		return
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), userID, req)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	response := map[string]any{
		"id":       updatedUser.ID.Hex(),
		"username": updatedUser.Username,
		"role":     updatedUser.Role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser handles the request to delete a user.
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *AdminHandler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.auditService.ListRecentEvents(r.Context(), 10) // Get last 10 events
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}
	type auditEventResponse struct {
		Timestamp string `json:"timestamp"`
		EventType string `json:"event_type"`
		ActorID   string `json:"actor_id"`
		Details   string `json:"details"`
	}
	response := make([]auditEventResponse, len(events))
	for i, event := range events {
		response[i] = auditEventResponse{
			// Format the time into a standard ISO 8601 string that JS can easily parse.
			Timestamp: event.Timestamp.Format(time.RFC3339),
			EventType: string(event.EventType),
			ActorID:   event.ActorID,
			Details:   event.Details,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
