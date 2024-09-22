package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultConnMaxIdleTime = 5 * time.Minute
	defaultConnMaxLifetime = 30 * time.Minute
	defaultMaxIdleConns    = 5
	defaultMaxOpenConns    = 25
)

// Option represents a functional option for configuring the database connection.
type Option func(*sqlx.DB)

// WithConnMaxIdleTime sets the maximum idle time for a connection in the pool.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(db *sqlx.DB) {
		db.SetConnMaxIdleTime(d)
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection before it is closed.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(db *sqlx.DB) {
		db.SetConnMaxLifetime(d)
	}
}

// WithMaxIdleConns sets the maximum number of idle connections allowed in the pool.
func WithMaxIdleConns(n int) Option {
	return func(db *sqlx.DB) {
		db.SetMaxIdleConns(n)
	}
}

// WithMaxOpenConns sets the maximum number of open connections in the pool.
func WithMaxOpenConns(n int) Option {
	return func(db *sqlx.DB) {
		db.SetMaxOpenConns(n)
	}
}

// New creates a new connection to the PostgreSQL database and applies the provided configuration options.
func New(ctx context.Context, dsn string, opts ...Option) (*sqlx.DB, error) {
	const op = "postgres.New"

	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	db.SetMaxIdleConns(defaultMaxIdleConns)
	db.SetMaxOpenConns(defaultMaxOpenConns)

	for _, opt := range opts {
		opt(db)
	}

	return db, nil
}
