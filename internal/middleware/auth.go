package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/aminshahid573/oauth2-provider/internal/models"
	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
)

// CtxUserKey is the key for storing the user object in the request context.
type CtxUserKey string

const UserKey CtxUserKey = "user"

// AuthMiddleware provides middleware for authentication.
type AuthMiddleware struct {
	logger         *slog.Logger
	sessionService *services.SessionService
	// *** THIS IS THE FIX ***
	// The middleware depends on the storage interface to fetch the user.
	userService storage.UserStore
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(logger *slog.Logger, sessionService *services.SessionService, userService storage.UserStore) *AuthMiddleware {
	return &AuthMiddleware{
		logger:         logger,
		sessionService: sessionService,
		userService:    userService,
	}
}

// RequireAuth is a middleware that ensures a user is authenticated.
// If not, it redirects them to the login page.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session_id")
		if err != nil { // No cookie found
			m.redirectToLogin(w, r)
			return
		}

		session, err := m.sessionService.GetSession(r.Context(), sessionCookie.Value)
		if err != nil { // Invalid session
			m.clearSessionCookie(w)
			m.redirectToLogin(w, r)
			return
		}

		// Fetch the full user object from the database using the storage interface
		user, err := m.userService.GetByID(r.Context(), session.UserID)
		if err != nil {
			m.clearSessionCookie(w)
			m.redirectToLogin(w, r)
			return
		}

		// Add the user to the request context for later handlers to use.
		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// Preserve the original URL as a 'return_to' query parameter.
	loginURL := "/login?return_to=" + url.QueryEscape(r.RequestURI)
	http.Redirect(w, r, loginURL, http.StatusSeeOther)
}

func (m *AuthMiddleware) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Deletes the cookie
	})
}

// GetUserFromContext retrieves the authenticated user from the request context.
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserKey).(*models.User)
	return user, ok
}
