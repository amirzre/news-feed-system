package repository

import (
	"context"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostRepository defines the contract for post data operations
type PostRepository interface {
	CreatePost(ctx context.Context, params *model.CreatePostParams) (*model.Post, error)
	GetPostByURL(ctx context.Context, url string) (*model.Post, error)
	GetPostByID(ctx context.Context, id int64) (*model.Post, error)
	UpdatePost(ctx context.Context, id int64, params *model.UpdatePostParams) (*model.Post, error)
	DeletePost(ctx context.Context, id int64) error
	CountPosts(ctx context.Context) (int64, error)
	CountPostsByCategory(ctx context.Context, category string) (int64, error)
	ListPostByCategory(ctx context.Context, params *model.ListPostsByCategoryParams) (*[]model.Post, error)
	ListPostsBySource(ctx context.Context, params *model.ListPostsBySourceParams) (*[]model.Post, error)
	SearchPost(ctx context.Context, params *model.SearchPostsParams) (*[]model.Post, error)
}

// Repository holds all repository implementations
type Repository struct {
	Post PostRepository
}

// New creates a new repository instance with all entity repositories
func New(db *pgxpool.Pool, redis *redis.Client, logger *logger.Logger, cacheTTL time.Duration) *Repository {
	return &Repository{
		Post: NewPostRepository(db, redis, logger, cacheTTL),
	}
}
