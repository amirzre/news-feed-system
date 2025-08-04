package database

import (
	"context"
	"fmt"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Database struct {
	PG    *pgxpool.Pool
	Redis *redis.Client
}

// NewDatabase creates new database connections
func NewDatabase(cfg *config.Config) (*Database, error) {
	pg, err := newPostgreSQLConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	rdb, err := newRedisConnection(cfg)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Database{
		PG:    pg,
		Redis: rdb,
	}, nil
}

// newPostgreSQLConnection creates a new PostgreSQL connection Pool
func newPostgreSQLConnection(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DatabasePool.MaxConns)
	poolConfig.MinConns = int32(cfg.DatabasePool.MinConns)
	poolConfig.MaxConnLifetime = cfg.DatabasePool.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DatabasePool.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.DatabasePool.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// newRedisConnection creates a new Redis connection
func newRedisConnection(cfg *config.Config) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     10,
		MinIdleConns: 3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return rdb, nil
}

// Close closes all database connections
func (db *Database) Close() {
	if db.PG != nil {
		db.PG.Close()
	}

	if db.Redis != nil {
		db.Redis.Close()
	}
}
