package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aminshahid573/authexa/internal/services"
)

// HealthHandler serves the liveness and readiness health check endpoints.
// Liveness indicates the process is running (Kubernetes livenessProbe).
// Readiness indicates all dependencies are reachable (Kubernetes readinessProbe).
//
// Response format follows draft-inadarei-api-health-check with per-component
// status reporting and RFC 9110 compliant HTTP status codes.
type HealthHandler struct {
	checker *services.HealthChecker
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(checker *services.HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// componentStatus represents the health of a single dependency.
type componentStatus struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// healthResponse is the JSON body returned by health endpoints.
type healthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]componentStatus `json:"components,omitempty"`
	Time       string                     `json:"time"`
}

// Live handles GET /health/live.
// Returns 200 OK if the Go process is running. No dependency checks.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status: "pass",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Ready handles GET /health/ready.
// Checks MongoDB and Redis connectivity. Returns 200 OK only if all
// dependencies are reachable. Returns 503 Service Unavailable with a
// per-component breakdown otherwise (RFC 9110 section 15.6.4).
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	results := h.checker.Check(r.Context())

	components := make(map[string]componentStatus, len(results))
	allHealthy := true

	for name, result := range results {
		cs := componentStatus{Status: "pass"}
		if result.Err != nil {
			cs.Status = "fail"
			cs.Detail = result.Err.Error()
			allHealthy = false
		}
		components[name] = cs
	}

	resp := healthResponse{
		Status:     "pass",
		Components: components,
		Time:       time.Now().UTC().Format(time.RFC3339),
	}

	httpStatus := http.StatusOK
	if !allHealthy {
		resp.Status = "fail"
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if httpStatus == http.StatusServiceUnavailable {
		w.Header().Set("Retry-After", "5")
	}
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}
