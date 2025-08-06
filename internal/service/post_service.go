package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/repository"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

// postService implements PostService interface
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

var (
	ErrPostExists = errors.New("post with this URL already exists")
)

// CreatePost creates a new post
func (s *postService) CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error) {
	start := time.Now()

	// TODO: check post exists with URL
	exists, err := s.repo.ExistsByURL(ctx, req.URL)
	if err != nil {
		s.logger.LogServiceOperation("post", "create", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to check post existence: %w", err)
	}

	if exists {
		s.logger.LogServiceOperation("post", "create", false, time.Since(start).Milliseconds())
		return nil, ErrPostExists
	}

	post, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.LogServiceOperation("post", "create", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to create post service: %w", err)
	}

	s.logger.LogServiceOperation("post", "create", true, time.Since(start).Milliseconds())

	return post, nil
}
