package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// CORS provides a middleware for handling Cross-Origin Resource Sharing.
// Debug logging is only enabled in non-production environments to prevent
// leaking internal routing information (RFC 6454, Fetch Living Standard).
func CORS(allowedOrigins []string, appEnv string) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodDelete, http.MethodPut},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Cookie"},
		AllowCredentials: true,
		Debug:            appEnv != "production",
	})

	return c.Handler
}
