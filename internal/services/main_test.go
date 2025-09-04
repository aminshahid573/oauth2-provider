package services

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/aminshahid573/oauth2-provider/internal/testutil"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	testMongoClient *mongo.Client
	testRedisClient *redis.Client
	testDB          *mongo.Database
)

// TestMain sets up the database connections before running tests and tears them down after.
func TestMain(m *testing.M) {
	testutil.SetRootPath()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config for integration tests: %v", err)
	}

	// Connect to MongoDB
	testMongoClient, err = mongo.Connect(options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		log.Fatalf("could not connect to mongodb for tests: %v", err)
	}
	testDB = testMongoClient.Database("oauth2_provider_test") // Use a separate test database

	// Connect to Redis
	testRedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       1, // Use a separate Redis DB for tests
	})
	if _, err := testRedisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("could not connect to redis for tests: %v", err)
	}

	// Run the tests
	exitCode := m.Run()

	// --- Teardown ---
	if err := testMongoClient.Disconnect(context.Background()); err != nil {
		log.Printf("error disconnecting from mongodb: %v", err)
	}
	if err := testRedisClient.Close(); err != nil {
		log.Printf("error closing redis client: %v", err)
	}

	os.Exit(exitCode)
}
