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

// App holds application-wide dependencies.
type App struct {
	Config       *config.Config
	Logger       *slog.Logger
	DataStore    *storage.DataStore
	SessionStore storage.SessionStore
	PKCEStore    storage.PKCEStore
	JWTManager   *utils.JWTManager // Add JWTManager
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup Structured Logger
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
	slog.SetDefault(logger)
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
	dataStore := &storage.DataStore{
		Client: mongodb.NewClientRepository(db),
		User:   mongodb.NewUserRepository(db),
		Token:  mongodb.NewTokenRepository(db),
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

	// Initialize Redis-backed stores
	sessionStore := redis.NewSessionRepository(redisClient)
	pkceStore := redis.NewPKCERepository(redisClient)
	logger.Info("Redis repositories initialized")

	// Initialize JWT Manager
	jwtManager := utils.NewJWTManager(cfg.JWT)
	logger.Info("JWT Manager initialized")

	// Create the main application container
	app := &App{
		Config:       cfg,
		Logger:       logger,
		DataStore:    dataStore,
		SessionStore: sessionStore,
		PKCEStore:    pkceStore,
		JWTManager:   jwtManager,
	}

	// We will replace this simple server with a proper router and handlers soon.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OAuth2 Provider is running.")
	})

	logger.Info("Starting server", "address", fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port))

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port), nil); err != nil {
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
