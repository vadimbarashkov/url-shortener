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

type Option func(*sqlx.DB)

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(db *sqlx.DB) {
		db.SetConnMaxIdleTime(d)
	}
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(db *sqlx.DB) {
		db.SetConnMaxLifetime(d)
	}
}

func WithMaxIdleConns(n int) Option {
	return func(db *sqlx.DB) {
		db.SetMaxIdleConns(n)
	}
}

func WithMaxOpenConns(n int) Option {
	return func(db *sqlx.DB) {
		db.SetMaxOpenConns(n)
	}
}

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
