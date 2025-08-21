// File: internal/server/router.go
package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/handlers"
	"github.com/aminshahid573/oauth2-provider/internal/services" // Import services
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"github.com/aminshahid573/oauth2-provider/web"
	"github.com/gorilla/csrf"
)

// AppDependencies holds the dependencies for the HTTP server.
type AppDependencies struct {
	Logger         *slog.Logger
	TemplateCache  utils.TemplateCache
	CSRFKey        string
	AuthService    *services.AuthService
	SessionService *services.SessionService
	ClientService  *services.ClientService // Added
	ScopeService   *services.ScopeService  // Added
	BaseURL        string
	AppEnv         string
}

// debugHeaders is a middleware for logging request headers.
func debugHeaders(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("[DEBUG HEADERS]",
			"method", r.Method,
			"path", r.URL.Path,
			"host", r.Host,
			"origin", r.Header.Get("Origin"),
			"referer", r.Header.Get("Referer"),
		)
		next.ServeHTTP(w, r)
	})
}

// NewRouter creates and configures the main application router.
func NewRouter(deps AppDependencies) http.Handler {
	deps.Logger.Info("Router configuration", "AppEnv", deps.AppEnv, "BaseURL", deps.BaseURL)

	mux := http.NewServeMux()

	// --- Static Files ---
	staticFS, _ := fs.Sub(web.Static, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// --- Frontend Handlers ---
	frontendHandler := handlers.NewFrontendHandler(deps.Logger, deps.TemplateCache, deps.AuthService, deps.SessionService, deps.ScopeService)
	mux.HandleFunc("GET /login", frontendHandler.LoginPage)
	mux.HandleFunc("POST /login", frontendHandler.Login)

	// --- OAuth2 Endpoints ---
	authHandler := handlers.NewAuthHandler(
		deps.Logger,
		deps.TemplateCache,
		deps.ClientService,
		deps.ScopeService,
		deps.SessionService,
	)
	mux.HandleFunc("GET /oauth2/authorize", authHandler.Authorize)
	// mux.HandleFunc("POST /oauth2/authorize", authHandler.HandleConsent) // Will be added next

	// --- Placeholder for admin dashboard ---
	mux.HandleFunc("GET /admin/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the dashboard!"))
	})

	// --- Middleware Configuration ---
	var handler http.Handler = mux

	// Conditionally apply CSRF middleware based on environment
	if deps.AppEnv == "development" {
		deps.Logger.Info("CSRF protection DISABLED for development environment")
	} else {
		csrfOpts := []csrf.Option{
			csrf.Secure(true),
			csrf.Path("/"),
			csrf.HttpOnly(true),
			csrf.SameSite(csrf.SameSiteLaxMode),
			csrf.TrustedOrigins([]string{deps.BaseURL}),
		}
		csrfMiddleware := csrf.Protect([]byte(deps.CSRFKey), csrfOpts...)
		handler = csrfMiddleware(handler)
		deps.Logger.Info("CSRF protection ENABLED for production environment")
	}

	// Apply the debug logger last, so it runs first.
	handler = debugHeaders(handler, deps.Logger)

	return handler
}
