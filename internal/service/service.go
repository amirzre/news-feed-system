package service

import (
	"context"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/repository"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

// PostService defines the contract for post business operations
type PostService interface {
	CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error)
	PostExists(ctx context.Context, url string) (bool, error)
	GetPostByID(ctx context.Context, id int64) (*model.Post, error)
	UpdatePost(ctx context.Context, id int64, req *model.UpdatePostParams) (*model.Post, error)
	DeletePost(ctx context.Context, id int64) error
}

// Service holds all service implementations
type Service struct {
	Post PostService
}

// New creates a new service instance with all entity services
func New(repo *repository.Repository, logger *logger.Logger) *Service {
	return &Service{
		Post: NewPostService(repo.Post, logger),
	}
}
