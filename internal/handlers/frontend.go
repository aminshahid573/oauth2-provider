package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// FrontendHandler handles requests for serving HTML pages.
type FrontendHandler struct {
	logger         *slog.Logger
	templateCache  utils.TemplateCache
	authService    *services.AuthService
	sessionService *services.SessionService
	// Add dependencies needed for device flow consent page
	tokenService  *services.TokenService
	clientService *services.ClientService
	scopeService  *services.ScopeService
}

// NewFrontendHandler creates a new FrontendHandler.
func NewFrontendHandler(
	logger *slog.Logger,
	templateCache utils.TemplateCache,
	authService *services.AuthService,
	sessionService *services.SessionService,
	tokenService *services.TokenService,
	clientService *services.ClientService,
	scopeService *services.ScopeService,
) *FrontendHandler {
	return &FrontendHandler{
		logger:         logger,
		templateCache:  templateCache,
		authService:    authService,
		sessionService: sessionService,
		tokenService:   tokenService,
		clientService:  clientService,
		scopeService:   scopeService,
	}
}

// LoginPage serves the user login page.
func (h *FrontendHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	returnTo := r.URL.Query().Get("return_to")
	data := map[string]any{
		"ReturnTo": returnTo,
	}
	h.templateCache.Render(w, r, "login.html", data)
}

// Login handles the submission of the login form.
func (h *FrontendHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	returnTo := r.PostForm.Get("return_to")

	validator := utils.NewValidator()
	validator.Check(utils.NotBlank(username), "username", "Username must not be empty.")
	validator.Check(utils.NotBlank(password), "password", "Password must not be empty.")

	if !validator.Valid() {
		data := map[string]any{"Username": username, "Validator": validator, "ReturnTo": returnTo}
		h.templateCache.Render(w, r, "login.html", data)
		return
	}

	user, err := h.authService.AuthenticateUser(r.Context(), username, password)
	if err != nil {
		if errors.Is(err, utils.ErrUnauthorized) {
			validator.AddError("credentials", "Invalid username or password.")
			data := map[string]any{"Username": username, "Validator": validator, "ReturnTo": returnTo}
			h.templateCache.Render(w, r, "login.html", data)
			return
		}
		utils.HandleError(w, r, h.logger, err)
		return
	}

	session, err := h.sessionService.CreateSession(r.Context(), user.ID)
	if err != nil {
		utils.HandleError(w, r, h.logger, err)
		return
	}

	isSecure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	if returnTo != "" && strings.HasPrefix(returnTo, "/") {
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// DeviceFlow handles the user-facing part of the device flow.
// It shows the code entry page (GET) and processes the code (POST).
func (h *FrontendHandler) DeviceFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templateCache.Render(w, r, "device.html", nil)
		return
	}

	// POST request logic
	if err := r.ParseForm(); err != nil {
		utils.HandleError(w, r, h.logger, utils.ErrBadRequest)
		return
	}
	userCode := strings.ToUpper(r.PostForm.Get("user_code"))

	// Validate the user code format (should be 10 characters)
	if len(userCode) != 10 {
		data := map[string]any{"Error": "Invalid code format."}
		h.templateCache.Render(w, r, "device.html", data)
		return
	}

	// Find the device code to get client and scope info
	token, err := h.tokenService.GetTokenByUserCode(r.Context(), userCode)
	if err != nil {
		h.logger.Warn("Failed to get token by user code", "user_code", userCode, "error", err)
		data := map[string]any{"Error": "Invalid or expired code."}
		h.templateCache.Render(w, r, "device.html", data)
		return
	}

	// Check if the token is still valid (not expired)
	if time.Now().After(token.ExpiresAt) {
		data := map[string]any{"Error": "Code has expired."}
		h.templateCache.Render(w, r, "device.html", data)
		return
	}

	// Redirect to the consent page with the user code
	redirectURL := "/oauth2/authorize/device?user_code=" + url.QueryEscape(userCode)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
