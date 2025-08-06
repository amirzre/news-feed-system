package repository

import (
	"context"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface{
	CreatePost(ctx context.Context, arg model.CreatePostParams) (model.Post, error)
}

type DBTX interface{
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Queries struct {
	db DBTX
}

// New creates a new Queries instance
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// WithTx creates a new Queries instance with transaction
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}
