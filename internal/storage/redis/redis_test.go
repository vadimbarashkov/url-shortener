package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/vadimbarashkov/url-shortener/internal/storage"
	my_redis "github.com/vadimbarashkov/url-shortener/internal/storage/redis"
	mock_redis "github.com/vadimbarashkov/url-shortener/internal/storage/redis/mock"
)

var ErrUnknown = errors.New("unknown error")

func TestURLStorage_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock_redis.NewMockClient(ctrl)
	urlStorage := my_redis.NewURLStorage(client)

	alias := "alias"
	url := "http://example.com"

	t.Run("return unknown error after setnx", func(t *testing.T) {
		client.
			EXPECT().
			SetNX(context.Background(), alias, url, time.Duration(0)).
			Return(redis.NewBoolResult(false, ErrUnknown))

		err := urlStorage.Add(context.Background(), alias, url)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnknown)
	})

	t.Run("return url exists error", func(t *testing.T) {
		client.
			EXPECT().
			SetNX(context.Background(), alias, url, time.Duration(0)).
			Return(redis.NewBoolResult(false, nil))

		err := urlStorage.Add(context.Background(), alias, url)

		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrURLExists)
	})

	t.Run("return no error", func(t *testing.T) {
		client.
			EXPECT().
			SetNX(context.Background(), alias, url, time.Duration(0)).
			Return(redis.NewBoolResult(true, nil))

		err := urlStorage.Add(context.Background(), alias, url)

		assert.NoError(t, err)
	})
}

func TestURLStorage_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock_redis.NewMockClient(ctrl)
	urlStorage := my_redis.NewURLStorage(client)

	alias := "alias"
	expectedURL := "http://example.com"

	t.Run("return url not found error", func(t *testing.T) {
		client.
			EXPECT().
			Get(context.Background(), alias).
			Return(redis.NewStringResult("", redis.Nil))

		url, err := urlStorage.Get(context.Background(), alias)

		assert.Equal(t, "", url)
		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrURLNotFound)
	})

	t.Run("return unknown error after get", func(t *testing.T) {
		client.
			EXPECT().
			Get(context.Background(), alias).
			Return(redis.NewStringResult("", ErrUnknown))

		url, err := urlStorage.Get(context.Background(), alias)

		assert.Equal(t, "", url)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnknown)
	})

	t.Run("return no error", func(t *testing.T) {
		client.
			EXPECT().
			Get(context.Background(), alias).
			Return(redis.NewStringResult("http://example.com", nil))

		url, err := urlStorage.Get(context.Background(), alias)

		assert.Equal(t, expectedURL, url)
		assert.NoError(t, err)
	})
}

func TestURLStorage_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock_redis.NewMockClient(ctrl)
	urlStorage := my_redis.NewURLStorage(client)

	alias := "alias"
	url := "http://example.com"

	t.Run("return unknown error after exists", func(t *testing.T) {
		client.
			EXPECT().
			Exists(context.Background(), alias).
			Return(redis.NewIntResult(0, ErrUnknown))

		err := urlStorage.Update(context.Background(), alias, url)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnknown)
	})

	t.Run("return url not found error", func(t *testing.T) {
		client.
			EXPECT().
			Exists(context.Background(), alias).
			Return(redis.NewIntResult(0, nil))

		err := urlStorage.Update(context.Background(), alias, url)

		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrURLNotFound)
	})

	t.Run("return unknown error after set", func(t *testing.T) {
		client.
			EXPECT().
			Exists(context.Background(), alias).
			Return(redis.NewIntResult(1, nil))

		client.
			EXPECT().
			Set(context.Background(), alias, url, time.Duration(0)).
			Return(redis.NewStatusResult("", ErrUnknown))

		err := urlStorage.Update(context.Background(), alias, url)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnknown)
	})

	t.Run("return no error", func(t *testing.T) {
		client.
			EXPECT().
			Exists(context.Background(), alias).
			Return(redis.NewIntResult(1, nil))

		client.
			EXPECT().
			Set(context.Background(), alias, url, time.Duration(0)).
			Return(redis.NewStatusResult("", nil))

		err := urlStorage.Update(context.Background(), alias, url)

		assert.NoError(t, err)
	})
}

func TestURLStorage_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock_redis.NewMockClient(ctrl)
	urlStorage := my_redis.NewURLStorage(client)

	alias := "alias"

	t.Run("return unknown error after del", func(t *testing.T) {
		client.
			EXPECT().
			Del(context.Background(), alias).
			Return(redis.NewIntResult(0, ErrUnknown))

		err := urlStorage.Delete(context.Background(), alias)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnknown)
	})

	t.Run("return url not found error", func(t *testing.T) {
		client.
			EXPECT().
			Del(context.Background(), alias).
			Return(redis.NewIntResult(0, nil))

		err := urlStorage.Delete(context.Background(), alias)

		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrURLNotFound)
	})

	t.Run("return no error", func(t *testing.T) {
		client.
			EXPECT().
			Del(context.Background(), alias).
			Return(redis.NewIntResult(1, nil))

		err := urlStorage.Delete(context.Background(), alias)

		assert.NoError(t, err)
	})
}
