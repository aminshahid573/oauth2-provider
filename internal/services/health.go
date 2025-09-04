package services

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// HealthChecker provides methods to check the health of application dependencies.
type HealthChecker struct {
	mongoClient *mongo.Client
	redisClient *redis.Client
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(mongoClient *mongo.Client, redisClient *redis.Client) *HealthChecker {
	return &HealthChecker{
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}

// Check performs a concurrent health check of all dependencies.
func (h *HealthChecker) Check(ctx context.Context) map[string]string {
	// Use a WaitGroup to run checks in parallel for speed.
	var wg sync.WaitGroup
	results := make(chan map[string]string, 2)
	wg.Add(2)

	// Check MongoDB
	go func() {
		defer wg.Done()
		status := "ok"
		// Use a short timeout for health checks.
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := h.mongoClient.Ping(checkCtx, readpref.Primary()); err != nil {
			status = "error: " + err.Error()
		}
		results <- map[string]string{"mongodb": status}
	}()

	// Check Redis
	go func() {
		defer wg.Done()
		status := "ok"
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if _, err := h.redisClient.Ping(checkCtx).Result(); err != nil {
			status = "error: " + err.Error()
		}
		results <- map[string]string{"redis": status}
	}()

	wg.Wait()
	close(results)

	// Consolidate results
	healthStatus := make(map[string]string)
	for res := range results {
		for k, v := range res {
			healthStatus[k] = v
		}
	}

	return healthStatus
}
