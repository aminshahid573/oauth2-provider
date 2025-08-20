// File: cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/storage/mongodb"
	"github.com/aminshahid573/oauth2-provider/internal/storage/redis"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Println("Configuration loaded successfully!")

	// Establish MongoDB connection
	mongoClient, err := mongodb.NewConnection(cfg.Mongo)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	fmt.Println("Successfully connected to MongoDB!")

	// Extract database name from URI for clarity
	dbName := getDBNameFromURI(cfg.Mongo.URI)
	db := mongoClient.Database(dbName)

	// Initialize MongoDB-backed stores
	clientStore := mongodb.NewClientRepository(db)
	userStore := mongodb.NewUserRepository(db)
	tokenStore := mongodb.NewTokenRepository(db)

	// Create the main data store container
	dataStore := &storage.DataStore{
		Client: clientStore,
		User:   userStore,
		Token:  tokenStore,
	}
	fmt.Println("MongoDB repositories initialized.")
	_ = dataStore // Use dataStore to avoid unused variable error for now

	// Establish Redis connection
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	fmt.Println("Successfully connected to Redis!")

	// --- Application startup logic will go here in the future ---
	fmt.Println("Application is ready to start. (Terminating for now)")
}

// getDBNameFromURI is a helper to extract the database name from the MongoDB URI.
func getDBNameFromURI(uri string) string {
	// A simple heuristic: find the last '/' and check for '?'
	if lastSlash := strings.LastIndex(uri, "/"); lastSlash != -1 {
		dbPart := uri[lastSlash+1:]
		if qIndex := strings.Index(dbPart, "?"); qIndex != -1 {
			return dbPart[:qIndex]
		}
		return dbPart
	}
	// Default if parsing fails
	return "oauth2_provider"
}
