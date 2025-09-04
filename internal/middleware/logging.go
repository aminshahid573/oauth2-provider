package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// StructuredLogger is a middleware that logs requests in a structured format.
func StructuredLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for static assets to reduce noise.
			if strings.HasPrefix(r.URL.Path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			// Use the shared responseWriter from util.go
			rw := newResponseWriter(w)

			// Serve the request
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			user, _ := GetUserFromContext(r)
			userID := "anonymous"
			if user != nil {
				userID = user.ID.Hex()
			}

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("duration", duration),
				slog.String("ip", GetClientIP(r)), // Use shared GetClientIP
				slog.String("user_agent", r.UserAgent()),
				slog.String("user_id", userID),
			)
		})
	}
}
