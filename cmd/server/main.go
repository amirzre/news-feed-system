package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/pkg/database"
	"github.com/amirzre/news-feed-system/pkg/logger"
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
}
