package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/amirzre/news-feed-system/internal/model"
)

// CreatePost creates a new post
func (q *Queries) CreatePost(ctx context.Context, arg model.CreatePostParams) (model.Post, error) {
	query := `
		INSERT INTO posts (title, description, content, url, source, category, image_url, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, content, url, source, category, image_url, published_at, created_at, updated_at
	`

	var post model.Post
	var publishedAt, createdAt, updatedAt sql.NullTime

	err := q.db.QueryRow(
		ctx,
		query,
		arg.Title,
		arg.Description,
		arg.Content,
		arg.URL,
		arg.Source,
		arg.Category,
		arg.ImageURL,
		arg.PublishedAt,
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
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		return model.Post{}, fmt.Errorf("Failed to create post: %w", err)
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}
	if createdAt.Valid {
		post.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		post.UpdatedAt = updatedAt.Time
	}

	return post, nil
}
