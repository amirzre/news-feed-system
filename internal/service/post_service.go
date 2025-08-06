package service

import (
	"context"
	"fmt"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/repository"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

type PostService interface {
	CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error)
}

type postService struct {
	repo   repository.PostRepository
	logger *logger.Logger
}

// NewPostService creates a new post service
func NewPostService(repo repository.PostRepository, logger *logger.Logger) PostService {
	return &postService{
		repo:   repo,
		logger: logger,
	}
}

// CreatePost creates a new post
func (s *postService) CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error) {
	start := time.Now()

	// TODO: check post exists with same URL

	post, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.LogServiceOperation("post", "create", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to create post service: %w", err)
	}

	s.logger.LogServiceOperation("post", "create", true, time.Since(start).Milliseconds())

	return post, nil
}
