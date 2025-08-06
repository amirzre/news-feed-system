package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/pkg/database"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/amirzre/news-feed-system/pkg/validator"
	"github.com/labstack/echo/v4"
)

const (
	appName    = "news-feed-system"
	appVersion = "1.0.0"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg)

	// Initialize database connection
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Error("Failed to initialize database connections", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Test database connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Health(ctx); err != nil {
		log.Error("Database health check failed", "error", err.Error())
		os.Exit(1)
	}

	log.Info("Database connections established successfully")

	// Initialize Echo server
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true
	e.Validator = validator.NewValidator()

	// Start server in a goroutine
	go func() {
		addr := cfg.ServerAddr()
		log.LogStartup(appName, appVersion, cfg.Server.Port)

		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.LogShutdown(appName, "received shutdown signal")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err.Error())
		os.Exit(1)
	}

	log.Info("Server shutdown completed")
}
