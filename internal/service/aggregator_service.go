package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

// aggregatorService implements AggregatorService interface
type aggregatorService struct {
	newsService NewsService
	postService PostService
	logger      *logger.Logger
	maxWorkers  int
}

// NewAggregatorService creates a new aggregator service
func NewAggregatorService(newsService NewsService, postService PostService, logger *logger.Logger) AggregatorService {
	return &aggregatorService{
		newsService: newsService,
		postService: postService,
		logger:      logger,
		maxWorkers:  5,
	}
}

// AggregateTopHeadlines aggregates top headlines from multiple categories
func (s *aggregatorService) AggregateTopHeadlines(ctx context.Context) (*model.AggregationResponse, error) {
	start := time.Now()
	s.logger.Info("Starting top headlines aggregation")

	categories := GetDefaultCategories()
	result := s.aggregateByCategories(ctx, categories, true)

	result.Duration = time.Since(start)
	s.logger.LogServiceOperation("aggregator", "aggregate_top_headlines", result.TotalErrors == 0, result.Duration.Milliseconds())

	s.logger.Info("Completed top headlines aggregation",
		"total_fetched", result.TotalFetched,
		"total_created", result.TotalCreated,
		"total_duplicates", result.TotalDuplicates,
		"total_errors", result.TotalErrors,
		"duration_ms", result.Duration.Milliseconds(),
	)

	return result, nil
}

// aggregateByCategories is the internal implementation for category-based aggregation
func (s *aggregatorService) aggregateByCategories(ctx context.Context, categories []string, useTopHeadlines bool) *model.AggregationResponse {
	result := &model.AggregationResponse{
		Categories: make(map[string]model.CategoryStats),
		Sources:    make(map[string]model.SourceStats),
		Errors:     []string{},
	}

	// Use a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, s.maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, category := range categories {
		wg.Add(1)
		go func(cat string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			categoryStats := s.processCategoryNews(ctx, cat, useTopHeadlines)

			mu.Lock()
			result.TotalFetched += categoryStats.Fetched
			result.TotalCreated += categoryStats.Created
			result.TotalDuplicates += categoryStats.Duplicates
			result.TotalErrors += categoryStats.Errors
			result.Categories[cat] = categoryStats
			mu.Unlock()
		}(category)
	}

	wg.Wait()

	return result
}

// processCategoryNews processes news for a single category
func (s *aggregatorService) processCategoryNews(ctx context.Context, category string, useTopHeadlines bool) model.CategoryStats {
	stats := model.CategoryStats{}

	var response *model.NewsAPIResponse
	var err error

	if useTopHeadlines {
		response, err = s.newsService.GetNewsByCategory(ctx, category, 50)
	} else {
		req := &model.NewsParams{
			Query:    category,
			Language: "en",
			PageSize: 50,
		}
		response, err = s.newsService.GetEverything(ctx, req)
	}

	if err != nil {
		s.logger.Error("Failed to fetch news for category", "category", category, "error", err.Error())
		stats.Errors++
		return stats
	}

	stats.Fetched = len(response.Articles)

	for _, article := range response.Articles {
		if article.Source.Name != "" {
			post, err := s.postService.CreatePostFromNewsAPI(ctx, &article)
			if err != nil {
				if errors.Is(err, ErrPostExists) {
					stats.Duplicates++
				} else {
					stats.Errors++
					s.logger.Warn("Failed to create post from article",
						"url", article.URL,
						"error", err.Error(),
					)
				}
				continue
			}

			if post != nil {
				stats.Created++
			}
		}
	}

	s.logger.Debug("Processed category news",
		"category", category,
		"fetched", stats.Fetched,
		"created", stats.Created,
		"duplicates", stats.Duplicates,
		"errors", stats.Errors,
	)

	return stats
}
