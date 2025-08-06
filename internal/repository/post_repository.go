package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// postRepository implements PostRepository interface with caching
type postRepository struct {
	db       *pgxpool.Pool
	redis    *redis.Client
	logger   *logger.Logger
	cacheTTL time.Duration
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *pgxpool.Pool, redis *redis.Client, logger *logger.Logger, cacheTTL time.Duration) PostRepository {
	return &postRepository{
		db:       db,
		redis:    redis,
		logger:   logger,
		cacheTTL: cacheTTL,
	}
}

// Create creates a new post in the database
func (r *postRepository) CreatePost(ctx context.Context, params *model.CreatePostParams) (*model.Post, error) {
	query := `
		INSERT INTO posts (title, description, content, url, source, category, image_url, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
	`
	var post model.Post
	var publishedAt sql.NullTime

	err := r.db.QueryRow(ctx, query,
		params.Title,
		params.Description,
		params.Content,
		params.URL,
		params.Source,
		params.Category,
		params.ImageURL,
		params.PublishedAt,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Description,
		&post.Content,
		&post.URL,
		&post.Source,
		&post.Category,
		&post.ImageURL,
		&publishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		r.logger.LogDBOperation("create", "posts", time.Since(time.Now()).Milliseconds(), err)
		return nil, fmt.Errorf("Failed to create post: %w", err)
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}

	r.invalidateListCaches(ctx)

	return &post, nil
}

// GetPostByURL retrieves a post by URL from database
func (r *postRepository) GetPostByURL(ctx context.Context, url string) (*model.Post, error) {
	start := time.Now()

	query := `
		SELECT id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
		FROM posts WHERE url = $1 LIMIT 1
	`

	var post model.Post
	var publishedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, url).Scan(
		&post.ID,
		&post.Title,
		&post.Description,
		&post.Content,
		&post.URL,
		&post.Source,
		&post.Category,
		&post.ImageURL,
		&publishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		r.logger.LogDBOperation("get_by_url", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("Failed to get post by url: %w", err)
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}

	r.logger.LogDBOperation("get_by_url", "posts", time.Since(start).Milliseconds(), nil)

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
