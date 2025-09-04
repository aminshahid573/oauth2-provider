package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/handlers"
	"github.com/aminshahid573/oauth2-provider/internal/middleware"
	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
	"github.com/aminshahid573/oauth2-provider/web"
	"github.com/gorilla/csrf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	IntrospectionHandler *handlers.IntrospectionHandler
	RevocationHandler    *handlers.RevocationHandler
	JWKSHandler          *handlers.JWKSHandler
	DiscoveryHandler     *handlers.DiscoveryHandler
	UserInfoHandler      *handlers.UserInfoHandler
	AdminHandler         *handlers.AdminHandler

	AllowedOrigins []string
	RateLimiter    *middleware.RateLimiter

	Metrics *middleware.MetricsMiddleware
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

	// The main router for all application routes.
	mux := http.NewServeMux()

	// --- Initialize Handlers and Middleware from Dependencies ---
	authMiddleware := middleware.NewAuthMiddleware(deps.Logger, deps.SessionService, deps.UserStore)
	frontendHandler := handlers.NewFrontendHandler(deps.Logger, deps.TemplateCache, deps.AuthService, deps.SessionService, deps.TokenService, deps.ClientService, deps.ScopeService)
	authHandler := handlers.NewAuthHandler(deps.Logger, deps.TemplateCache, deps.ClientService, deps.ScopeService, deps.TokenService)

	// == Route Definitions ==

	// --- Static File Server ---
	staticFS, _ := fs.Sub(web.Static, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// --- Public User-Facing Routes (No Auth Required) ---
	mux.HandleFunc("GET /login", frontendHandler.LoginPage)
	mux.HandleFunc("POST /login", frontendHandler.Login)
	mux.HandleFunc("POST /logout", frontendHandler.Logout)

	// --- Protected User-Facing Routes (Login Required) ---
	mux.Handle("/device", authMiddleware.RequireAuth(http.HandlerFunc(frontendHandler.DeviceFlow)))
	mux.Handle("/oauth2/authorize", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.AuthorizeFlow)))

	deviceConsentHandler := http.HandlerFunc(authHandler.DeviceConsentFlow)
	mux.Handle("GET /oauth2/authorize/device", authMiddleware.RequireAuth(deviceConsentHandler))

	deviceConsentPostHandler := http.HandlerFunc(authHandler.HandleDeviceConsent)
	mux.Handle("POST /oauth2/authorize/device/consent", authMiddleware.RequireAuth(deviceConsentPostHandler))

	// --- Admin UI Routes (Login + Admin Role Required) ---
	adminUI := http.NewServeMux()
	adminUI.HandleFunc("GET /dashboard", frontendHandler.AdminDashboard)
	adminUI.HandleFunc("GET /clients", frontendHandler.AdminClientsPage)

	protectedAdminUI := authMiddleware.RequireAuth(authMiddleware.RequireAdmin(adminUI))
	mux.Handle("/admin/", http.StripPrefix("/admin", protectedAdminUI))

	// --- Admin API Routes (Login + Admin Role Required) ---
	adminAPI := http.NewServeMux()
	adminAPI.HandleFunc("GET /clients", deps.AdminHandler.ListClients)
	adminAPI.HandleFunc("POST /clients", deps.AdminHandler.CreateClient)
	adminAPI.HandleFunc("GET /clients/{clientID}", deps.AdminHandler.GetClient)
	adminAPI.HandleFunc("PUT /clients/{clientID}", deps.AdminHandler.UpdateClient)
	adminAPI.HandleFunc("DELETE /clients/{clientID}", deps.AdminHandler.DeleteClient)
	protectedAdminAPI := authMiddleware.RequireAuth(authMiddleware.RequireAdmin(adminAPI))
	mux.Handle("/api/admin/", http.StripPrefix("/api/admin", protectedAdminAPI))

	// --- Public OAuth2 API & Metadata Endpoints ---
	mux.HandleFunc("POST /oauth2/device_authorization", authHandler.DeviceAuthorization)
	mux.HandleFunc("POST /oauth2/introspect", deps.IntrospectionHandler.Introspect)
	mux.HandleFunc("POST /oauth2/revoke", deps.RevocationHandler.Revoke)
	mux.HandleFunc("GET /.well-known/jwks.json", deps.JWKSHandler.ServeJWKS)
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", deps.DiscoveryHandler.ServeDiscoveryDocument)
	mux.HandleFunc("/oauth2/userinfo", deps.UserInfoHandler.GetUserInfo)

	// The /token endpoint has its own specific rate limiter.
	tokenHandler := deps.RateLimiter.PerClient(http.HandlerFunc(authHandler.Token))
	mux.Handle("POST /oauth2/token", tokenHandler)

	// --- Public Metrics Endpoint ---
	mux.Handle("GET /metrics", promhttp.Handler())

	// == Central Middleware Chain ==
	var handler http.Handler = mux

	handler = deps.Metrics.Wrap(handler)

	// Conditionally apply CSRF middleware based on environment
	if deps.AppEnv == "development" {
		deps.Logger.Info("CSRF protection DISABLED for development environment")
		// Don't apply CSRF middleware in development
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

	// Apply the Global rate limiter to ALL requests.
	handler = deps.RateLimiter.Global(handler)

	handler = middleware.SecurityHeaders(handler)

	handler = middleware.CORS(deps.AllowedOrigins)(handler)

	// Apply the debug logger last, so it runs first.
	handler = debugHeaders(handler, deps.Logger)

	topLevelMux := http.NewServeMux()
	topLevelMux.Handle("GET /metrics", promhttp.Handler())
	topLevelMux.Handle("/", handler)

	return topLevelMux
}
