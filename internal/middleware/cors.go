package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// CORS provides a middleware for handling Cross-Origin Resource Sharing.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodDelete, http.MethodPut},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Cookie"},
		AllowCredentials: true,
		// Enable Debugging for development environment if needed
		Debug: true,
	})

	return c.Handler
}
