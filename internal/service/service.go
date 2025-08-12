package service

import (
	"context"

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
type AggregatorService interface{}

// Service holds all service implementations
type Service struct {
	Post       PostService
	News       NewsService
	Aggregator AggregatorService
}

// New creates a new service instance with all entity services
func New(repo *repository.Repository, logger *logger.Logger, cfg *config.Config) *Service {
	postSvc := NewPostService(repo.Post, logger)
	newsSvc := NewNewsService(cfg, logger)
	aggregatorSvc := NewAggregatorService(newsSvc, postSvc, logger)

	return &Service{
		Post:       postSvc,
		News:       newsSvc,
		Aggregator: aggregatorSvc,
	}
}
