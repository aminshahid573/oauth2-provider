package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/services"
)

// HealthHandler serves the health check endpoint.
type HealthHandler struct {
	checker *services.HealthChecker
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(checker *services.HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// ServeHTTP is the handler for the health check endpoint.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	healthStatus := h.checker.Check(r.Context())

	// Determine the overall status. If any check fails, the service is unhealthy.
	overallStatus := "ok"
	httpStatusCode := http.StatusOK
	for _, status := range healthStatus {
		if !strings.HasPrefix(status, "ok") {
			overallStatus = "error"
			httpStatusCode = http.StatusServiceUnavailable
			break
		}
	}

	response := map[string]any{
		"status":       overallStatus,
		"dependencies": healthStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	json.NewEncoder(w).Encode(response)
}
