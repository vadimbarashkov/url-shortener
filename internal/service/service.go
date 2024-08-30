package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ErrMaxRetriesExceeded = errors.New("maximum retries exceeded for generating short code")

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

func (s *URLService) ShortenURL(ctx context.Context, originalURL string) (*models.URL, error) {
	const op = "service.URLService.ShortenURL"
	const maxRetries = 5

	for i := 0; i < maxRetries; i++ {
		shortCode, err := gonanoid.New(s.shortCodeLength)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate short code: %w", op, err)
		}

		url, err := s.repo.Create(ctx, shortCode, originalURL)
		if err != nil {
			if errors.Is(err, database.ErrShortCodeExists) {
				continue
			}

			return nil, fmt.Errorf("%s: failed to shorten url: %w", op, err)
		}

		return url, nil
	}

	return nil, fmt.Errorf("%s: %w", op, ErrMaxRetriesExceeded)
}

func (s *URLService) ResolveShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	const op = "service.URLService.ResolveShortCode"

	url, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to resolve short code: %w", op, err)
	}

	return url, nil
}

func (s *URLService) ModifyURL(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	const op = "service.URLService.ModifyURL"

	url, err := s.repo.Update(ctx, shortCode, originalURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to modify url: %w", op, err)
	}

	return url, nil
}

func (s *URLService) DeactivateURL(ctx context.Context, shortCode string) error {
	const op = "service.URLService.DeactivateURL"

	err := s.repo.Delete(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("%s: failed to deactivate url: %w", op, err)
	}

	return nil
}

func (s *URLService) GetURLStats(ctx context.Context, shortCode string) (*models.URL, error) {
	const op = "service.URLService.GetURLStats"

	url, err := s.repo.GetStats(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get url stats: %w", op, err)
	}

	return url, nil
}
