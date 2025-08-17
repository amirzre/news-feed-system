package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
		return nil, fmt.Errorf("failed to create post: %w", err)
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
		return nil, fmt.Errorf("failed to get post by url: %w", err)
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}

	r.logger.LogDBOperation("get_by_url", "posts", time.Since(start).Milliseconds(), nil)

	return &post, nil
}

// GetPostByID retrieves a post by ID with caching
func (r *postRepository) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	start := time.Now()
	cacheKey := fmt.Sprintf("post:id:%d", id)

	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var post model.Post
		if err := json.Unmarshal([]byte(cached), &post); err == nil {
			r.logger.LogCacheOperation("get", cacheKey, true)
			return &post, nil
		}
	}
	r.logger.LogCacheOperation("get", cacheKey, false)

	query := `
		SELECT id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
		FROM posts WHERE id = $1 LIMIT 1
	`
	var post model.Post
	var publishedAt sql.NullTime

	err = r.db.QueryRow(ctx, query, id).Scan(
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
		r.logger.LogDBOperation("get_by_id", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}
	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}

	r.logger.LogDBOperation("get_by_id", "posts", time.Since(start).Milliseconds(), nil)

	if postJson, err := json.Marshal(post); err == nil {
		r.redis.Set(ctx, cacheKey, postJson, r.cacheTTL).Err()
		r.logger.LogCacheOperation("set", cacheKey, false)
	}

	return &post, nil
}

// UpdatePost updates a post in the database
func (r *postRepository) UpdatePost(ctx context.Context, id int64, params *model.UpdatePostParams) (*model.Post, error) {
	start := time.Now()

	query := `
		UPDATE posts 
		SET title = $2, description = $3, content = $4, category = $5, image_url = $6, updated_at = NOW()
		WHERE id = $1 
		RETURNING id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
	`
	var post model.Post
	var publishedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id,
		params.Title,
		params.Description,
		params.Content,
		params.Category,
		params.ImageURL,
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
		r.logger.LogDBOperation("update", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}

	r.logger.LogDBOperation("update", "posts", time.Since(start).Milliseconds(), nil)

	r.invalidatePostCaches(ctx, id)
	r.invalidateListCaches(ctx)

	return &post, nil
}

// DeletePost deletes a post from the database
func (r *postRepository) DeletePost(ctx context.Context, id int64) error {
	start := time.Now()

	query := `DELETE FROM posts WHERE id = $1`

	post, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.LogDBOperation("delete", "posts", time.Since(start).Milliseconds(), err)
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if post.RowsAffected() == 0 {
		return fmt.Errorf("post with id %d not found", id)
	}

	r.logger.LogDBOperation("delete", "posts", time.Since(start).Milliseconds(), nil)

	r.invalidatePostCaches(ctx, id)
	r.invalidateListCaches(ctx)

	return nil
}

// ListPosts retrieves posts with pagination
func (r *postRepository) ListPosts(ctx context.Context, params *model.PostListParams) ([]model.Post, error) {
	start := time.Now()

	limit := params.Limit
	offset := (params.Page - 1) * params.Limit

	var posts []model.Post
	var err error

	switch {
	case params.Search != nil && *params.Search != "":
		posts, err = r.SearchPosts(ctx, &model.SearchPostsParams{
			BasePostListParams: model.BasePostListParams{Limit: limit, Offset: offset},
			Query:              *params.Search,
		})
	case params.Category != nil && *params.Category != "":
		posts, err = r.ListPostsByCategory(ctx, &model.ListPostsByCategoryParams{
			BasePostListParams: model.BasePostListParams{Limit: limit, Offset: offset},
			Category:           *params.Category,
		})
	case params.Source != nil && *params.Source != "":
		posts, err = r.ListPostsBySource(ctx, &model.ListPostsBySourceParams{
			BasePostListParams: model.BasePostListParams{Limit: limit, Offset: offset},
			Source:             *params.Source,
		})
	default:
		cacheKey := fmt.Sprintf("posts:list:%d:%d", params.Page, params.Limit)
		cached, cacheErr := r.redis.Get(ctx, cacheKey).Result()
		if cacheErr == nil {
			if err := json.Unmarshal([]byte(cached), &posts); err == nil {
				r.logger.LogCacheOperation("get", cacheKey, true)
				return posts, nil
			}
		}
		r.logger.LogCacheOperation("get", cacheKey, false)

		query := `
			SELECT id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
			FROM posts ORDER BY published_at DESC LIMIT $1 OFFSET $2
		`
		rows, err := r.db.Query(ctx, query, limit, offset)
		if err != nil {
			r.logger.LogDBOperation("list", "posts", time.Since(start).Milliseconds(), err)
			return nil, fmt.Errorf("failed to list posts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var post model.Post
			var publishedAt sql.NullTime

			err = rows.Scan(
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

				return nil, fmt.Errorf("failed to scan post: %w", err)
			}
			if publishedAt.Valid {
				post.PublishedAt = &publishedAt.Time
			}

			posts = append(posts, post)
		}

		if err := rows.Err(); err != nil {
			r.logger.LogDBOperation("list", "posts", time.Since(start).Milliseconds(), err)
			return nil, fmt.Errorf("failed to iterate posts: %w", err)
		}

		if err == nil {
			if postsJSON, jsonErr := json.Marshal(posts); jsonErr == nil {
				r.redis.Set(ctx, cacheKey, postsJSON, r.cacheTTL).Err()
				r.logger.LogCacheOperation("set", cacheKey, false)
			}
		}
	}

	if err != nil {
		r.logger.LogDBOperation("list", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	r.logger.LogDBOperation("list", "posts", time.Since(start).Milliseconds(), nil)

	return posts, nil
}

// ListPostsByCategory retrieves posts by category
func (r *postRepository) ListPostsByCategory(ctx context.Context, params *model.ListPostsByCategoryParams) ([]model.Post, error) {
	start := time.Now()

	query := `
		SELECT id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
		FROM posts WHERE category = $1 ORDER BY published_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, params.Category, params.Limit, params.Offset)
	if err != nil {
		r.logger.LogDBOperation("list_by_category", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to list posts by category: %w", err)
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		var publishedAt sql.NullTime

		err = rows.Scan(
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
			r.logger.LogDBOperation("list_by_category", "posts", time.Since(start).Milliseconds(), err)
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		if publishedAt.Valid {
			post.PublishedAt = &publishedAt.Time
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		r.logger.LogDBOperation("list_by_category", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to iterate posts by category: %w", err)
	}

	r.logger.LogDBOperation("list_by_category", "posts", time.Since(start).Milliseconds(), nil)

	return posts, nil
}

// ListPostsBySource retrieves posts by source
func (r *postRepository) ListPostsBySource(ctx context.Context, params *model.ListPostsBySourceParams) ([]model.Post, error) {
	start := time.Now()

	query := `
		SELECT id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
		FROM posts WHERE source = $1 ORDER BY published_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, params.Source, params.Limit, params.Offset)
	if err != nil {
		r.logger.LogDBOperation("list_by_source", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to list posts by source: %w", err)
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		var publishedAt sql.NullTime

		err = rows.Scan(
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
			r.logger.LogDBOperation("list_by_source", "posts", time.Since(start).Milliseconds(), err)
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		if publishedAt.Valid {
			post.PublishedAt = &publishedAt.Time
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		r.logger.LogDBOperation("list_by_source", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to iterate posts by source: %w", err)
	}

	r.logger.LogDBOperation("list_by_source", "posts", time.Since(start).Milliseconds(), nil)

	return posts, nil
}

// SearchPosts searches posts
func (r *postRepository) SearchPosts(ctx context.Context, params *model.SearchPostsParams) ([]model.Post, error) {
	start := time.Now()

	query := `
		SELECT id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
		FROM posts 
		WHERE title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%'
		ORDER BY published_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, params.Query, params.Limit, params.Offset)
	if err != nil {
		r.logger.LogDBOperation("search", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		var publishedAt sql.NullTime

		err = rows.Scan(
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
			r.logger.LogDBOperation("search", "posts", time.Since(start).Milliseconds(), err)
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		if publishedAt.Valid {
			post.PublishedAt = &publishedAt.Time
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		r.logger.LogDBOperation("search", "posts", time.Since(start).Milliseconds(), err)
		return nil, fmt.Errorf("failed to iterate search results: %w", err)
	}

	r.logger.LogDBOperation("search", "posts", time.Since(start).Milliseconds(), nil)

	return posts, nil
}

// CountPosts counts all posts
func (r *postRepository) CountPosts(ctx context.Context) (int64, error) {
	start := time.Now()
	cacheKey := "posts:count"

	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var count int64
		if err := json.Unmarshal([]byte(cached), &count); err == nil {
			r.logger.LogCacheOperation("get", cacheKey, true)
			return count, nil
		}
	}
	r.logger.LogCacheOperation("get", cacheKey, false)

	query := `SELECT COUNT(*) FROM posts`

	var count int64
	err = r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.LogDBOperation("count", "posts", time.Since(start).Milliseconds(), err)
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}

	r.logger.LogDBOperation("count", "posts", time.Since(start).Milliseconds(), nil)

	if countJSON, err := json.Marshal(count); err == nil {
		r.redis.Set(ctx, cacheKey, countJSON, r.cacheTTL).Err()
		r.logger.LogCacheOperation("set", cacheKey, false)
	}

	return count, nil
}

// CountByCategory returns the number of posts in a category
func (r *postRepository) CountPostsByCategory(ctx context.Context, category string) (int64, error) {
	start := time.Now()

	query := `SELECT COUNT(*) FROM posts WHERE category = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, category).Scan(&count)
	if err != nil {
		r.logger.LogDBOperation("count_by_category", "posts", time.Since(start).Milliseconds(), err)
		return 0, fmt.Errorf("failed to count posts by category: %w", err)
	}

	r.logger.LogDBOperation("count_by_category", "posts", time.Since(start).Milliseconds(), nil)

	return count, nil
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
