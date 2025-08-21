package handlers

import (
	"log/slog"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/services" // Import services
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// FrontendHandler handles requests for serving HTML pages.
type FrontendHandler struct {
	logger         *slog.Logger
	templateCache  utils.TemplateCache
	authService    *services.AuthService
	sessionService *services.SessionService
}

// NewFrontendHandler creates a new FrontendHandler.
func NewFrontendHandler(logger *slog.Logger, templateCache utils.TemplateCache, authService *services.AuthService, sessionService *services.SessionService) *FrontendHandler {
	return &FrontendHandler{
		logger:         logger,
		templateCache:  templateCache,
		authService:    authService,
		sessionService: sessionService,
	}
}

// LoginPage serves the user login page.
func (h *FrontendHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.templateCache.Render(w, r, "login.html", nil)
}
