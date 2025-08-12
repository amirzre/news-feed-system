package service

import (
	"context"
	"errors"
	"fmt"
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

// AggregateByCategories aggregates news from specific categories
func (s *aggregatorService) AggregateByCategories(ctx context.Context, categories []string) (*model.AggregationResponse, error) {
	start := time.Now()
	s.logger.Info("Starting category-based aggregation", "categories", categories)

	result := s.aggregateByCategories(ctx, categories, true)
	result.Duration = time.Since(start)

	s.logger.LogServiceOperation("aggregator", "aggregate_by_categories", result.TotalErrors == 0, result.Duration.Milliseconds())

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

// aggregateBySources is the internal implementation for source-based aggregation
func (s *aggregatorService) aggregateBySources(ctx context.Context, sources []string) *model.AggregationResponse {
	result := &model.AggregationResponse{
		Categories: make(map[string]model.CategoryStats),
		Sources:    make(map[string]model.SourceStats),
		Errors:     []string{},
	}

	batchSize := 3
	semaphore := make(chan struct{}, s.maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < len(sources); i += batchSize {
		end := i + batchSize
		if end > len(sources) {
			end = len(sources)
		}

		batch := sources[i:end]

		wg.Add(1)
		go func(sourceBatch []string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			batchStats := s.processSourceNews(ctx, sourceBatch)

			mu.Lock()
			result.TotalFetched += batchStats.TotalFetched
			result.TotalCreated += batchStats.TotalCreated
			result.TotalDuplicates += batchStats.TotalDuplicates
			result.TotalErrors += batchStats.TotalErrors

			for k, v := range batchStats.Sources {
				result.Sources[k] = v
			}
			result.Errors = append(result.Errors, batchStats.Errors...)
			mu.Unlock()
		}(batch)
	}

	wg.Wait()
	return result
}

// processSourceNews processes news from a batch of sources
func (s *aggregatorService) processSourceNews(ctx context.Context, sources []string) *model.AggregationResponse {
	result := &model.AggregationResponse{
		Sources: make(map[string]model.SourceStats),
		Errors:  []string{},
	}

	response, err := s.newsService.GetNewsBySources(ctx, sources, 100)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch news for sources %v: %v", sources, err)
		s.logger.Error(errorMsg)
		result.Errors = append(result.Errors, errorMsg)
		result.TotalErrors++
		return result
	}

	result.TotalFetched = len(response.Articles)

	sourceStats := make(map[string]model.SourceStats)
	for _, source := range sources {
		sourceStats[source] = model.SourceStats{}
	}

	for _, article := range response.Articles {
		sourceName := article.Source.Name

		post, err := s.postService.CreatePostFromNewsAPI(ctx, &article)
		if err != nil {
			if err.Error() == "post with this URL already exists" {
				result.TotalDuplicates++
				if stats, ok := sourceStats[sourceName]; ok {
					stats.Duplicates++
					sourceStats[sourceName] = stats
				}
			} else {
				result.TotalErrors++
				if stats, ok := sourceStats[sourceName]; ok {
					stats.Errors++
					sourceStats[sourceName] = stats
				}
				s.logger.Warn("Failed to create post from article",
					"url", article.URL,
					"source", sourceName,
					"error", err.Error(),
				)
			}
			continue
		}

		if post != nil {
			result.TotalCreated++
			if stats, ok := sourceStats[sourceName]; ok {
				stats.Created++
				stats.Fetched++
				sourceStats[sourceName] = stats
			}
		}
	}

	result.Sources = sourceStats

	s.logger.Debug("Processed source news",
		"sources", sources,
		"fetched", result.TotalFetched,
		"created", result.TotalCreated,
		"duplicates", result.TotalDuplicates,
		"errors", result.TotalErrors,
	)

	return result
}
