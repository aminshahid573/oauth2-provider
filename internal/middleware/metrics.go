package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// responseWriter is a wrapper for http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	// Default to 200 OK if WriteHeader is not called.
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware holds the Prometheus metrics.
type MetricsMiddleware struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewMetricsMiddleware creates and registers the Prometheus metrics.
func NewMetricsMiddleware() *MetricsMiddleware {
	m := &MetricsMiddleware{}

	m.requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "code"},
	)

	m.requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets, // Default buckets are a good starting point.
		},
		[]string{"method", "path"},
	)

	return m
}

// Wrap wraps an http.Handler to record metrics.
func (m *MetricsMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		// Serve the request
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path // For simplicity, we use the full path.

		// Record the metrics
		m.requestDuration.WithLabelValues(r.Method, path).Observe(duration)
		m.requestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rw.statusCode)).Inc()
	})
}
