package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	return m
}

// Wrap wraps an http.Handler to record metrics.
func (m *MetricsMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Use the shared responseWriter from util.go
		rw := NewResponseWriter(w)

		// Serve the request
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path

		// Record the metrics
		m.requestDuration.WithLabelValues(r.Method, path).Observe(duration)
		m.requestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rw.StatusCode())).Inc()
	})
}
