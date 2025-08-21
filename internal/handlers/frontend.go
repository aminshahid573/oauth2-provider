package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// FrontendHandler handles requests for serving HTML pages.
type FrontendHandler struct {
	logger         *slog.Logger
	templateCache  utils.TemplateCache
	authService    *services.AuthService
	sessionService *services.SessionService
	scopeService   *services.ScopeService
}

// NewFrontendHandler creates a new FrontendHandler.
func NewFrontendHandler(logger *slog.Logger, templateCache utils.TemplateCache, authService *services.AuthService, sessionService *services.SessionService, scopeService *services.ScopeService) *FrontendHandler {
	return &FrontendHandler{
		logger:         logger,
		templateCache:  templateCache,
		authService:    authService,
		sessionService: sessionService,
		scopeService:   scopeService,
	}
}

// LoginPage serves the user login page.
func (h *FrontendHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Extract the return_to parameter from the query string.
	returnTo := r.URL.Query().Get("return_to")

	// Pass it to the template.
	data := map[string]any{
		"ReturnTo": returnTo,
	}
	h.templateCache.Render(w, r, "login.html", data)
}

// Login handles the submission of the login form.
func (h *FrontendHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Parse the form data.
	if err := r.ParseForm(); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	// 2. Validate the input.
	validator := utils.NewValidator()
	validator.Check(utils.NotBlank(username), "username", "Username must not be empty.")
	validator.Check(utils.NotBlank(password), "password", "Password must not be empty.")

	if !validator.Valid() {
		// If validation fails, re-render the login page with error messages.
		data := map[string]any{
			"Username":  username, // Pre-fill the username
			"Validator": validator,
		}
		h.templateCache.Render(w, r, "login.html", data)
		return
	}

	// 3. Authenticate the user.
	user, err := h.authService.AuthenticateUser(r.Context(), username, password)
	if err != nil {
		// Check if it's a simple unauthorized error.
		if errors.Is(err, utils.ErrUnauthorized) {
			validator.AddError("credentials", "Invalid username or password.")
			data := map[string]any{
				"Username":  username,
				"Validator": validator,
			}
			h.templateCache.Render(w, r, "login.html", data)
			return
		}
		// For any other error, it's an internal server issue.
		utils.HandleError(w, r, h.logger, err)
		return
	}

	// 4. Create a new session.
	session, err := h.sessionService.CreateSession(r.Context(), user.ID)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	// 5. Set the session cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil, // True in production
		SameSite: http.SameSiteLaxMode,
	})

	// 6. Redirect the user.
	returnTo := r.PostForm.Get("return_to")

	// Security check: Ensure the return_to URL is a local path to prevent open redirect vulnerabilities.
	if returnTo != "" && strings.HasPrefix(returnTo, "/") {
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
		return
	}

	// If no valid return_to is provided, redirect to the default dashboard.
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
