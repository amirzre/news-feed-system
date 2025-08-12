package service

import "github.com/amirzre/news-feed-system/pkg/logger"

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
