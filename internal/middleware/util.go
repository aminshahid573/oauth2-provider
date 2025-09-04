package middleware

import (
	"net"
	"net/http"
	"strings"
)

// responseWriter is a wrapper for http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new responseWriter wrapper.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	// Default to 200 OK if WriteHeader is not called.
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code before writing it to the underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetClientIP retrieves the real client IP address, considering proxies.
func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	ip, _, _ = net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	return ip
}
