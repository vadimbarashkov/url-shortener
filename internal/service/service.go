package service

import (
	"context"

	"github.com/vadimbarashkov/url-shortener/internal/models"
)

type URLRepository interface {
	Create(ctx context.Context, shortCode, originalURL string) (*models.URL, error)
	GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	Update(ctx context.Context, shortCode, originalURL string) (*models.URL, error)
	Delete(ctx context.Context, shortCode string) error
	GetStats(ctx context.Context, shortCode string) (*models.URL, error)
}

type URLService struct {
	repo            URLRepository
	shortCodeLength int
}

func NewURLService(repo URLRepository, shortCodeLength int) *URLService {
	return &URLService{
		repo:            repo,
		shortCodeLength: shortCodeLength,
	}
}
