package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vadimbarashkov/url-shortener/internal/storage"
)

//go:generate mockgen -source=redis.go -destination=mocks/redis.go
type Client interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type urlStorage struct {
	client Client
}

func NewURLStorage(client Client) storage.URLStorage {
	return &urlStorage{
		client: client,
	}
}

func (s *urlStorage) Set(ctx context.Context, alias, url string) error {
	wasSet, err := s.client.SetNX(ctx, alias, url, 0).Result()
	if err != nil {
		return fmt.Errorf("urlStorage.Set: failed to setnx url: %w", err)
	}

	if !wasSet {
		return storage.ErrURLExists
	}

	return nil
}

func (s *urlStorage) Get(ctx context.Context, alias string) (string, error) {
	url, err := s.client.Get(ctx, alias).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("urlStorage.Get: failed to get url: %w", err)
	}

	return url, nil
}

func (s *urlStorage) Update(ctx context.Context, alias, url string) error {
	exists, err := s.client.Exists(ctx, alias).Result()
	if err != nil {
		return fmt.Errorf("urlStorage.Update: failed to exists url: %w", err)
	}

	if exists == 0 {
		return storage.ErrURLNotFound
	}

	err = s.client.Set(ctx, alias, url, 0).Err()
	if err != nil {
		return fmt.Errorf("urlStorage.Update: failed to set url: %w", err)
	}

	return nil
}

func (s *urlStorage) Delete(ctx context.Context, alias string) error {
	deleted, err := s.client.Del(ctx, alias).Result()
	if err != nil {
		return fmt.Errorf("urlStorage.Delete: failed to delete url: %w", err)
	}

	if deleted == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
