package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

type MockURLRepository struct {
	mock.Mock
}

func (r *MockURLRepository) Create(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	args := r.Called(ctx, shortCode, originalURL)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}

func (r *MockURLRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	args := r.Called(ctx, shortCode)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}

func (r *MockURLRepository) Update(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	args := r.Called(ctx, shortCode, originalURL)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}

func (r *MockURLRepository) Delete(ctx context.Context, shortCode string) error {
	args := r.Called(ctx, shortCode)
	return args.Error(0)
}

func (r *MockURLRepository) GetStats(ctx context.Context, shortCode string) (*models.URL, error) {
	args := r.Called(ctx, shortCode)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}
