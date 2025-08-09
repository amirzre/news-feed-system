package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/repository"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/jackc/pgx/v5"
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
	ErrPostExists    = errors.New("post with this URL already exists")
	ErrPostIDInvalid = errors.New("post ID is invalid")
	ErrPostNotFound  = errors.New("post not found")
)

// CreatePost creates a new post
func (s *postService) CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error) {
	start := time.Now()

	exists, err := s.PostExists(ctx, req.URL)
	if err != nil {
		s.logger.LogServiceOperation("post", "create", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to check post existence: %w", err)
	}

	if exists {
		s.logger.LogServiceOperation("post", "create", false, time.Since(start).Milliseconds())
		return nil, ErrPostExists
	}

	post, err := s.repo.CreatePost(ctx, req)
	if err != nil {
		s.logger.LogServiceOperation("post", "create", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to create post service: %w", err)
	}

	s.logger.LogServiceOperation("post", "create", true, time.Since(start).Milliseconds())

	return post, nil
}

// GetPostByID retrieves a post by ID
func (s *postService) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	start := time.Now()

	if id <= 0 {
		s.logger.LogServiceOperation("post", "get_by_id", false, time.Since(start).Milliseconds())
		return nil, ErrPostIDInvalid
	}

	post, err := s.repo.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.LogServiceOperation("post", "get_by_id", false, time.Since(start).Milliseconds())
			return nil, ErrPostNotFound
		}

		s.logger.LogServiceOperation("post", "get_by_id", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to get post: %w", err)
	}

	s.logger.LogServiceOperation("post", "get_by_id", true, time.Since(start).Milliseconds())

	return post, nil
}

// ListPosts retrieves posts with pagination and filtering
func (s *postService) ListPosts(ctx context.Context, req *model.PostListParams) (*model.PostListResponse, error) {
	start := time.Now()

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	posts, err := s.repo.ListPosts(ctx, req)
	if err != nil {
		s.logger.LogServiceOperation("post", "list", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to list posts: %w", err)
	}

	var total int64
	if req.Category != nil && *req.Category != "" {
		total, err = s.repo.CountPostsByCategory(ctx, *req.Category)
	} else {
		total, err = s.repo.CountPosts(ctx)
	}

	if err != nil {
		s.logger.LogServiceOperation("post", "list", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to count posts: %w", err)
	}

	pagination := model.CalculatePagination(req.Page, req.Limit, total)

	response := &model.PostListResponse{
		Posts:      posts,
		Pagination: pagination,
	}

	s.logger.LogServiceOperation("post", "list", true, time.Since(start).Milliseconds())

	return response, nil
}

// UpdatePost updates an existing post
func (s *postService) UpdatePost(ctx context.Context, id int64, req *model.UpdatePostParams) (*model.Post, error) {
	start := time.Now()

	if id <= 0 {
		s.logger.LogServiceOperation("post", "update", false, time.Since(start).Milliseconds())
		return nil, ErrPostIDInvalid
	}

	_, err := s.repo.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.LogServiceOperation("post", "update", false, time.Since(start).Milliseconds())
			return nil, ErrPostNotFound
		}

		s.logger.LogServiceOperation("post", "update", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to check post existence: %w", err)
	}

	post, err := s.repo.UpdatePost(ctx, id, req)
	if err != nil {
		s.logger.LogServiceOperation("post", "update", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("Failed to update post: %w", err)
	}

	s.logger.LogServiceOperation("post", "update", true, time.Since(start).Milliseconds())

	return post, nil
}

// DeletePost deletes a post
func (s *postService) DeletePost(ctx context.Context, id int64) error {
	start := time.Now()

	if id <= 0 {
		s.logger.LogServiceOperation("post", "delete", false, time.Since(start).Milliseconds())
		return ErrPostIDInvalid
	}

	_, err := s.repo.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.LogServiceOperation("post", "delete", false, time.Since(start).Milliseconds())
			return ErrPostNotFound
		}

		s.logger.LogServiceOperation("post", "delete", false, time.Since(start).Milliseconds())
		return fmt.Errorf("Failed to check post existence: %w", err)
	}

	err = s.repo.DeletePost(ctx, id)
	if err != nil {
		s.logger.LogServiceOperation("post", "delete", false, time.Since(start).Milliseconds())
		return fmt.Errorf("Failed to delete post: %w", err)
	}

	s.logger.LogServiceOperation("post", "delete", true, time.Since(start).Milliseconds())

	return nil
}

// PostExists checks if a post with the given URL already exists
func (s *postService) PostExists(ctx context.Context, url string) (bool, error) {
	start := time.Now()

	_, err := s.repo.GetPostByURL(ctx, url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.LogServiceOperation("post", "exists", false, time.Since(start).Milliseconds())
			return false, nil
		}
	}

	s.logger.LogServiceOperation("post", "exists", true, time.Since(start).Milliseconds())

	return true, nil
}
