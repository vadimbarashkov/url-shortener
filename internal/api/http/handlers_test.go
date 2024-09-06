package http

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

type MockURLService struct {
	mock.Mock
}

func (s *MockURLService) ShortenURL(ctx context.Context, originalURL string) (*models.URL, error) {
	args := s.Called(ctx, originalURL)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}

func (s *MockURLService) ResolveShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	args := s.Called(ctx, shortCode)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}

func (s *MockURLService) ModifyURL(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	args := s.Called(ctx, shortCode, originalURL)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}

func (s *MockURLService) DeactivateURL(ctx context.Context, shortCode string) error {
	args := s.Called(ctx, shortCode)
	return args.Error(0)
}

func (s *MockURLService) GetURLStats(ctx context.Context, shortCode string) (*models.URL, error) {
	args := s.Called(ctx, shortCode)
	url, _ := args.Get(0).(*models.URL)
	return url, args.Error(1)
}
