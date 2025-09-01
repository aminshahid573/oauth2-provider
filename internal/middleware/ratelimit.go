// File: internal/middleware/ratelimit.go
package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RateLimiter provides middleware for rate limiting requests.
type RateLimiter struct {
	limiter *redis_rate.Limiter
	cfg     config.RateLimitConfig
	logger  *slog.Logger
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(redisClient *redis.Client, cfg config.RateLimitConfig, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		limiter: redis_rate.NewLimiter(redisClient),
		cfg:     cfg,
		logger:  logger,
	}
}

// Global is a middleware that applies a rate limit based on the client's IP address.
func (rl *RateLimiter) Global(next http.Handler) http.Handler {
	if !rl.cfg.GlobalEnabled {
		rl.logger.Info("Global rate limiting is DISABLED")
		return next
	}

	rl.logger.Info("Global rate limiting is ENABLED",
		"rps", rl.cfg.GlobalRPS,
		"burst", rl.cfg.GlobalBurst)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		key := "global:" + ip

		limit := redis_rate.Limit{
			Rate:   rl.cfg.GlobalRPS,
			Period: time.Second,
			Burst:  rl.cfg.GlobalBurst,
		}

		rl.logger.Info("Rate limit check",
			"key", key,
			"rate", limit.Rate,
			"burst", limit.Burst,
			"ip", ip,
			"path", r.URL.Path)

		res, err := rl.limiter.Allow(r.Context(), key, limit)
		if err != nil {
			rl.logger.Error("Rate limiter failed", "error", err, "key", key)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		rl.logger.Info("Rate limit result",
			"key", key,
			"allowed", res.Allowed,
			"remaining", res.Remaining,
			"resetAfter", res.ResetAfter)

		setRateLimitHeaders(w, res)

		if res.Allowed == 0 {
			rl.logger.Warn("Rate limit exceeded", "key", key, "ip", ip)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// PerClient is a middleware that applies a rate limit based on the `client_id` in the request body.
func (rl *RateLimiter) PerClient(next http.Handler) http.Handler {
	if !rl.cfg.TokenEnabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		clientID := r.PostForm.Get("client_id")
		if clientID == "" {
			next.ServeHTTP(w, r)
			return
		}

		limit := redis_rate.Limit{
			Rate:   rl.cfg.TokenRPS,
			Period: time.Second,
			Burst:  rl.cfg.TokenBurst,
		}
		res, err := rl.limiter.Allow(r.Context(), "client:"+clientID, limit)
		if err != nil {
			rl.logger.Error("Rate limiter failed", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return

		}

		// We still set the headers here, they might be overwritten by the Global limiter, which is fine.
		setRateLimitHeaders(w, res)

		if res.Allowed == 0 {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	// For local testing, this will likely be "127.0.0.1" or "::1"
	ip := r.Header.Get("X-Forwarded-For")
	ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	if ip != "" {
		fmt.Printf("Using X-Forwarded-For IP: %s\n", ip)
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		fmt.Printf("Using X-Real-Ip IP: %s\n", ip)
		return ip
	}

	ip, _, _ = net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	fmt.Printf("Using RemoteAddr IP: %s\n", ip)
	return ip
}

func setRateLimitHeaders(w http.ResponseWriter, res *redis_rate.Result) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(res.Limit.Rate))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(res.ResetAfter).Unix(), 10))
	if res.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
	}
}
