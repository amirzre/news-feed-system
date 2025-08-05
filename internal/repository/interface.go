package repository

import "github.com/jackc/pgx/v5"

type Querier interface{}

type DBTX interface{}

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
