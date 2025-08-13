package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

// SetupAggregationJobs registers all aggregation jobs.
func SetupAggregationJobs(scheduler service.SchedulerService, aggregator service.AggregatorService, log *logger.Logger) {
	// Top headlines every 30 minutes
	scheduler.AddJob("top-headlines", 30*time.Minute, func(ctx context.Context) error {
		log.Info("Running scheduled top headlines aggregation")
		result, err := aggregator.AggregateTopHeadlines(ctx)
		if err != nil {
			return fmt.Errorf("failed to aggregate top headline news job: %w", err)
		}

		log.Info("Top headlines aggregation completed",
			"fetched", result.TotalFetched,
			"created", result.TotalCreated,
			"duplicates", result.TotalDuplicates,
			"errors", result.TotalErrors,
		)

		return nil
	})

	// Category-based aggregation every 2 hours
	scheduler.AddJob("category-aggregation", 2*time.Hour, func(ctx context.Context) error {
		log.Info("Running scheduled category aggregation")
		categories := service.GetDefaultCategories()
		result, err := aggregator.AggregateByCategories(ctx, categories)
		if err != nil {
			return fmt.Errorf("failed to aggregate category news job: %w", err)
		}

		log.Info("Category aggregation completed",
			"fetched", result.TotalFetched,
			"created", result.TotalCreated,
			"duplicates", result.TotalDuplicates,
			"errors", result.TotalErrors,
		)

		return nil
	})

	// Source-based aggregation every 4 hours
	scheduler.AddJob("source-aggregation", 4*time.Hour, func(ctx context.Context) error {
		log.Info("Running scheduled source aggregation")
		sources := service.GetDefaultSources()
		result, err := aggregator.AggregateBySources(ctx, sources)
		if err != nil {
			return fmt.Errorf("failed to aggregate source news job: %w", err)
		}

		log.Info("Source aggregation completed",
			"fetched", result.TotalFetched,
			"created", result.TotalCreated,
			"duplicates", result.TotalDuplicates,
			"errors", result.TotalErrors,
		)

		return nil
	})

	log.Info("Aggregation jobs configured successfully")
}
