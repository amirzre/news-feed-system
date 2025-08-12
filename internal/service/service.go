package service

import (
	"context"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/repository"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

// PostService defines the contract for post business operations
type PostService interface {
	CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error)
	PostExists(ctx context.Context, url string) (bool, error)
	GetPostByID(ctx context.Context, id int64) (*model.Post, error)
	ListPosts(ctx context.Context, req *model.PostListParams) (*model.PostListResponse, error)
	UpdatePost(ctx context.Context, id int64, req *model.UpdatePostParams) (*model.Post, error)
	DeletePost(ctx context.Context, id int64) error
	CreatePostFromNewsAPI(ctx context.Context, article *model.NewsAPIArticleParams) (*model.Post, error)
}

// NewsService defines the contract for news business operations
type NewsService interface {
	GetTopHeadlines(ctx context.Context, req *model.NewsParams) (*model.NewsAPIResponse, error)
	GetEverything(ctx context.Context, req *model.NewsParams) (*model.NewsAPIResponse, error)
	GetNewsByCategory(ctx context.Context, category string, pageSize int) (*model.NewsAPIResponse, error)
	GetNewsBySources(ctx context.Context, sources []string, pageSize int) (*model.NewsAPIResponse, error)
}

// AggregatorService defines the contract for aggregator business operations
type AggregatorService interface {
	AggregateTopHeadlines(ctx context.Context) (*model.AggregationResponse, error)
	AggregateByCategories(ctx context.Context, categories []string) (*model.AggregationResponse, error)
	AggregateBySources(ctx context.Context, sources []string) (*model.AggregationResponse, error)
	AggregateAll(ctx context.Context) (*model.AggregationResponse, error)
}

// SchedulerService defines the contract for scheduler business operations
type SchedulerService interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
	AddJob(name string, interval time.Duration, job func(context.Context) error)
	RemoveJob(name string)
}

// Service holds all service implementations
type Service struct {
	Post       PostService
	News       NewsService
	Aggregator AggregatorService
	Scheduler  SchedulerService
}

// New creates a new service instance with all entity services
func New(repo *repository.Repository, logger *logger.Logger, cfg *config.Config) *Service {
	postSvc := NewPostService(repo.Post, logger)
	newsSvc := NewNewsService(cfg, logger)
	aggregatorSvc := NewAggregatorService(newsSvc, postSvc, logger)
	schedulerSvc := NewSchedulerService(logger)

	return &Service{
		Post:       postSvc,
		News:       newsSvc,
		Aggregator: aggregatorSvc,
		Scheduler:  schedulerSvc,
	}
}
