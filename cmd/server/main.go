// File: cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/storage/mongodb"
	"github.com/aminshahid573/oauth2-provider/internal/storage/redis"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// --- Setup Structured Logger ---
	var logLevel slog.Level
	switch cfg.Log.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger) // Set as the default logger
	logger.Info("Configuration loaded and logger initialized")

	// Establish MongoDB connection
	mongoClient, err := mongodb.NewConnection(cfg.Mongo)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())
	logger.Info("Successfully connected to MongoDB")

	dbName := getDBNameFromURI(cfg.Mongo.URI)
	db := mongoClient.Database(dbName)

	// Initialize MongoDB-backed stores
	clientStore := mongodb.NewClientRepository(db)
	userStore := mongodb.NewUserRepository(db)
	tokenStore := mongodb.NewTokenRepository(db)

	dataStore := &storage.DataStore{
		Client: clientStore,
		User:   userStore,
		Token:  tokenStore,
	}
	logger.Info("MongoDB repositories initialized")

	// Establish Redis connection
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	logger.Info("Successfully connected to Redis")

	// We will replace this with a real router and handlers later.
	http.HandleFunc("/test-error", func(w http.ResponseWriter, r *http.Request) {
		// Simulate fetching a client that doesn't exist
		_, err := dataStore.Client.GetByClientID(context.Background(), "non-existent-client")
		if err != nil {
			utils.HandleError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This should not be reached"))
	})

	logger.Info("Starting server", "address", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))

	// We will replace this simple server later
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), nil); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

func getDBNameFromURI(uri string) string {
	if lastSlash := strings.LastIndex(uri, "/"); lastSlash != -1 {
		dbPart := uri[lastSlash+1:]
		if qIndex := strings.Index(dbPart, "?"); qIndex != -1 {
			return dbPart[:qIndex]
		}
		return dbPart
	}
	return "oauth2_provider"
}
