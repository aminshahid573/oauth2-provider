package handlers

import (
	"errors"
	"log/slog"
	"net/http"

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
	h.templateCache.Render(w, r, "login.html", nil)
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
	// For now, we'll redirect to a placeholder dashboard.
	// Later, this will redirect back to the OAuth2 flow.
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
