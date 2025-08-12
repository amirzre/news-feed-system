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
}

// NewsService defines the contract for news business operations
type NewsService interface{
	GetTopHeadlines(ctx context.Context, params *model.NewsParams) (*model.NewsAPIResponse, error)
}

// Service holds all service implementations
type Service struct {
	Post PostService
	News NewsService
}

// New creates a new service instance with all entity services
func New(repo *repository.Repository, logger *logger.Logger, cfg *config.Config) *Service {
	return &Service{
		Post: NewPostService(repo.Post, logger),
		News: NewNewsService(cfg, logger),
	}
}
