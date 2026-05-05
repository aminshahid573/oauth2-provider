package services

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// ComponentResult holds the outcome of a single dependency health check.
type ComponentResult struct {
	Err error
}

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
// Each dependency is probed with a 2-second timeout to prevent slow
// checks from blocking the readiness response.
func (h *HealthChecker) Check(ctx context.Context) map[string]ComponentResult {
	type entry struct {
		name   string
		result ComponentResult
	}

	var wg sync.WaitGroup
	results := make(chan entry, 2)
	wg.Add(2)

	// Check MongoDB
	go func() {
		defer wg.Done()
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		err := h.mongoClient.Ping(checkCtx, readpref.Primary())
		results <- entry{name: "mongodb", result: ComponentResult{Err: err}}
	}()

	// Check Redis
	go func() {
		defer wg.Done()
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		_, err := h.redisClient.Ping(checkCtx).Result()
		results <- entry{name: "redis", result: ComponentResult{Err: err}}
	}()

	wg.Wait()
	close(results)

	healthStatus := make(map[string]ComponentResult, 2)
	for e := range results {
		healthStatus[e.name] = e.result
	}

	return healthStatus
}
