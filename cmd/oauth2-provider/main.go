package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aminshahid573/oauth2-provider/internal/config"
	"github.com/aminshahid573/oauth2-provider/internal/handlers"
)

func setupLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func main() {
	logger := setupLogger()

	cfg := config.MustLoad(logger)

	mainMux := http.NewServeMux()

	mainMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	mainMux.HandleFunc("GET /login", handlers.LoginGetHandler)
	mainMux.HandleFunc("POST /login", handlers.LoginPostHandler)

	
	mainMux.HandleFunc("GET /authorize", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is the authorization endpoint"))
	})

	server := &http.Server{
		Handler:      mainMux,
		Addr:         cfg.Addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Server is starting with HTTPS", "address", cfg.HTTPServer.Addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("Server is ready to accept connections")

	<-done

	logger.Info("Received shutdown signal")
	logger.Info("Shutting down server gracefully")

	shutDownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutDownContext); err != nil {
		logger.Error("Failed to shutdown the server", "error", err)
	} else {
		logger.Info("Server shutdown gracefully")
	}
}
