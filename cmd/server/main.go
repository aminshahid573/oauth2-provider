// File: cmd/server/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/aminshahid573/oauth2-provider/internal/server"
	"github.com/aminshahid573/oauth2-provider/internal/services"
	"github.com/aminshahid573/oauth2-provider/internal/storage"
	"github.com/aminshahid573/oauth2-provider/internal/storage/mongodb"
	"github.com/aminshahid573/oauth2-provider/internal/storage/redis"
	"github.com/aminshahid573/oauth2-provider/internal/utils"
)

// App holds all application-wide dependencies, acting as a dependency injection container.
type App struct {
	Config         *config.Config
	Logger         *slog.Logger
	TemplateCache  utils.TemplateCache
	DataStore      *storage.DataStore
	SessionStore   storage.SessionStore
	PKCEStore      storage.PKCEStore
	JWTManager     *utils.JWTManager
	ClientService  *services.ClientService
	AuthService    *services.AuthService
	TokenService   *services.TokenService
	PKCEService    *services.PKCEService
	SessionService *services.SessionService
	ScopeService   *services.ScopeService
}

func main() {
	// The run function is used to allow for deferred calls to be executed before os.Exit.
	if err := run(); err != nil {
		slog.Error("application startup error", "error", err)
		os.Exit(1)
	}
}

// run initializes and starts the application.
func run() error {
	// --- Configuration ---
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// --- Logger ---
	logger := initLogger(cfg.Log.Level)
	slog.SetDefault(logger)
	logger.Info("configuration loaded and logger initialized")

	// --- Database Connections ---
	mongoClient, err := mongodb.NewConnection(cfg.Mongo)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			logger.Error("error disconnecting from MongoDB", "error", err)
		}
	}()
	logger.Info("successfully connected to MongoDB")

	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redisClient.Close()
	logger.Info("successfully connected to Redis")

	// --- Initialize Stores ---
	db := mongoClient.Database("oauth2_provider")
	dataStore := &storage.DataStore{
		Client: mongodb.NewClientRepository(db),
		User:   mongodb.NewUserRepository(db),
		Token:  mongodb.NewTokenRepository(db),
	}
	sessionStore := redis.NewSessionRepository(redisClient)
	pkceStore := redis.NewPKCERepository(redisClient)
	logger.Info("data stores initialized")

	// --- Initialize Services & Utilities ---
	jwtManager := utils.NewJWTManager(cfg.JWT)
	clientService := services.NewClientService(dataStore.Client)
	authService := services.NewAuthService(dataStore.User)
	tokenService := services.NewTokenService(jwtManager, dataStore.Token)
	pkceService := services.NewPKCEService(pkceStore)
	sessionService := services.NewSessionService(sessionStore)
	scopeService := services.NewScopeService()
	logger.Info("core services initialized")

	// --- Template Cache ---
	templateCache, err := utils.NewTemplateCache()
	if err != nil {
		return fmt.Errorf("failed to create template cache: %w", err)
	}
	logger.Info("template cache created successfully")

	// --- Assemble App Container ---
	app := &App{
		Config:         cfg,
		Logger:         logger,
		TemplateCache:  templateCache,
		DataStore:      dataStore,
		SessionStore:   sessionStore,
		PKCEStore:      pkceStore,
		JWTManager:     jwtManager,
		ClientService:  clientService,
		AuthService:    authService,
		TokenService:   tokenService,
		PKCEService:    pkceService,
		SessionService: sessionService,
		ScopeService:   scopeService,
	}

	// --- HTTP Server ---
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port),
		Handler:      server.NewRouter(app.ToServerDependencies()),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// --- Graceful Shutdown ---
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutdown signal received", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			shutdownError <- err
		}

		logger.Info("server gracefully stopped")
		shutdownError <- nil
	}()

	logger.Info("starting server", "address", srv.Addr)

	// Start the server. If it fails with an error other than ErrServerClosed,
	// it's a critical error.
	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed to start: %w", err)
	}

	// Wait for the graceful shutdown to complete.
	err = <-shutdownError
	if err != nil {
		return fmt.Errorf("error during server shutdown: %w", err)
	}

	logger.Info("application shut down complete")
	return nil
}

// initLogger initializes and returns a structured logger.
func initLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
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
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}

// ToServerDependencies maps the main App struct to the dependency struct required by the server/router.
// This is a clean way to pass only the necessary dependencies to the HTTP layer.
func (a *App) ToServerDependencies() server.AppDependencies {
	return server.AppDependencies{
		Logger:         a.Logger,
		TemplateCache:  a.TemplateCache,
		CSRFKey:        a.Config.CSRF.AuthKey,
		AuthService:    a.AuthService,
		SessionService: a.SessionService,
		ClientService:  a.ClientService,
		ScopeService:   a.ScopeService,
		BaseURL:        a.Config.BaseURL,
		AppEnv:         a.Config.AppEnv,
	}
}
