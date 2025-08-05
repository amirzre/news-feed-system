package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostRepository interface {
	Create(ctx context.Context, req *model.CreatePostParams) (*model.Post, error)
}

type postRepository struct {
	db       Querier
	redis    *redis.Client
	logger   *logger.Logger
	cacheTTL time.Duration
}

// NewPostRepository creates a new post repository
func NewPostRepository(pool *pgxpool.Pool, redis *redis.Client, logger *logger.Logger, cacheTTL time.Duration) PostRepository {
	return &postRepository{
		db:       New(pool),
		redis:    redis,
		logger:   logger,
		cacheTTL: cacheTTL,
	}
}

// Create creates a new post
func (r *postRepository) Create(ctx context.Context, req *model.CreatePostParams) (*model.Post, error) {
	defer func() {
		r.logger.LogDBOperation("create", "posts", int64(time.Since(time.Now()).Milliseconds()), nil)
	}()

	params := model.CreatePostParams{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		URL:         req.URL,
		Source:      req.Source,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		PublishedAt: req.PublishedAt,
	}

	post, err := r.db.CreatePost(ctx, params)
	if err != nil {
		r.logger.LogDBOperation("create", "posts", int64(time.Since(time.Now()).Milliseconds()), err)
	}

	r.invalidateListCaches(ctx)

	return &post, nil
}

// Helper methods for cache invalidation
func (r *postRepository) invalidatePostCaches(ctx context.Context, id int64) {
	cacheKey := fmt.Sprintf("post:id:%d", id)
	r.redis.Del(ctx, cacheKey).Err()
	r.logger.LogCacheOperation("delete", cacheKey, false)
}

func (r *postRepository) invalidateListCaches(ctx context.Context) {
	keys, err := r.redis.Keys(ctx, "posts:list:*").Result()
	if err == nil && len(keys) > 0 {
		r.redis.Del(ctx, keys...).Err()
		r.logger.LogCacheOperation("delete_pattern", "posts:list:*", false)
	}

	r.redis.Del(ctx, "posts:count").Err()
	r.logger.LogCacheOperation("delete", "posts:count", false)
}
