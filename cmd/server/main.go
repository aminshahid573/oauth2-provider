package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/aminshahid573/oauth2-provider/internal/storage/mongodb"
	"github.com/aminshahid573/oauth2-provider/internal/storage/redis"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("Configuration loaded successfully!")

	mongoClient, err := mongodb.NewConnection(cfg.Mongo)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	fmt.Println("Successfully connected to MongoDB!")

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
