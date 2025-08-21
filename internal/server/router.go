// File: internal/server/router.go
package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/handlers"
	"github.com/aminshahid573/oauth2-provider/internal/middleware"
	"github.com/aminshahid573/oauth2-provider/internal/services" // Import services
	"github.com/aminshahid573/oauth2-provider/internal/storage"
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
	ClientService  *services.ClientService
	ScopeService   *services.ScopeService
	TokenService   *services.TokenService
	UserStore      storage.UserStore
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

	// --- Initialize Middleware ---
	authMiddleware := middleware.NewAuthMiddleware(deps.Logger, deps.SessionService, deps.UserStore)

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
		deps.TokenService,
	)

	// The AuthorizeFlow handler is now protected by the auth middleware.
	mux.Handle("/oauth2/authorize", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.AuthorizeFlow)))

	// The /token endpoint is for clients, so it's NOT protected by session auth.
	mux.HandleFunc("POST /oauth2/token", authHandler.Token)

	// --- Placeholder for admin dashboard (now also protected) ---
	mux.Handle("/admin/dashboard", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the dashboard!"))
	})))

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
