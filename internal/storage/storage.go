package storage

import (
	"context"
	"errors"
)

var (
	ErrURLExists   = errors.New("url exists")
	ErrURLNotFound = errors.New("url not found")
)

type URLStorage interface {
	Add(ctx context.Context, alias, url string) error
	Get(ctx context.Context, alias string) (string, error)
	Update(ctx context.Context, alias, url string) error
	Delete(ctx context.Context, alias string) error
}
