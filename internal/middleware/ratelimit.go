package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
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
	rl.logger.Info("Global rate limiting is ENABLED", "rps", rl.cfg.GlobalRPS, "burst", rl.cfg.GlobalBurst)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r) // Use shared GetClientIP
		key := "global:" + ip
		limit := redis_rate.Limit{
			Rate:   rl.cfg.GlobalRPS,
			Period: time.Second,
			Burst:  rl.cfg.GlobalBurst,
		}
		rl.logger.Debug("Rate limit check", "key", key, "rate", limit.Rate, "burst", limit.Burst, "ip", ip, "path", r.URL.Path)

		res, err := rl.limiter.Allow(r.Context(), key, limit)
		if err != nil {
			rl.logger.Error("Rate limiter failed", "error", err, "key", key)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		rl.logger.Debug("Rate limit result", "key", key, "allowed", res.Allowed, "remaining", res.Remaining)

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
			rl.logger.Error("Per-client rate limiter failed", "error", err, "client_id", clientID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		setRateLimitHeaders(w, res)
		if res.Allowed == 0 {
			rl.logger.Warn("Per-client rate limit exceeded", "client_id", clientID)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// setRateLimitHeaders adds standard rate limit headers to the response.
func setRateLimitHeaders(w http.ResponseWriter, res *redis_rate.Result) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(res.Limit.Rate))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(res.ResetAfter).Unix(), 10))
	if res.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
	}
}
