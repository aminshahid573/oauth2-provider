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
	AuthService    *services.AuthService    // Add this
	SessionService *services.SessionService // Add this
}

// NewRouter creates and configures the main application router.
func NewRouter(deps AppDependencies) http.Handler {
	mux := http.NewServeMux()

	// --- Static Files ---
	staticFS, _ := fs.Sub(web.Static, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// --- Frontend Handlers ---
	frontendHandler := handlers.NewFrontendHandler(deps.Logger, deps.TemplateCache, deps.AuthService, deps.SessionService)
	mux.HandleFunc("GET /login", frontendHandler.LoginPage)
	// We will add the POST handler next
	// mux.HandleFunc("POST /login", frontendHandler.Login)

	// --- Middleware ---
	csrfMiddleware := csrf.Protect(
		[]byte(deps.CSRFKey),
		csrf.Secure(false), // Set to true in production with HTTPS
		csrf.Path("/"),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteLaxMode),
	)

	return csrfMiddleware(mux)
}
