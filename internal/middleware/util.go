package middleware

import (
	"net"
	"net/http"
	"strings"
)

// responseWriter is a wrapper for http.ResponseWriter to capture the status code.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new responseWriter wrapper.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	// Default to 200 OK if WriteHeader is not called.
	return &ResponseWriter{w, http.StatusOK}
}

func (rw *ResponseWriter) StatusCode() int {
	return rw.statusCode
}

// WriteHeader captures the status code before writing it to the underlying ResponseWriter.
func (rw *ResponseWriter) WriteHeader(code int) {
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
