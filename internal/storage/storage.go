package storage

import (
	"context"
	"errors"
)

var (
	ErrURLExists   = errors.New("url exists")
	ErrURLNotFound = errors.New("url not found")
)

//go:generate mockgen -source=storage.go -destination=mock/storage.go
type URLStorage interface {
	Add(ctx context.Context, alias, url string) error
	Get(ctx context.Context, alias string) (string, error)
	Update(ctx context.Context, alias, url string) error
	Delete(ctx context.Context, alias string) error
}
