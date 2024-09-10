package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// ErrMaxRetriesExceeded is returned when the maximum number of retries for generating a short code is exceeded.
var ErrMaxRetriesExceeded = errors.New("maximum retries exceeded for generating short code")

// URLRepository defines the interface for working with URLs at the business logic layer.
type URLRepository interface {
	// Create inserts a new shortened URL into the repository.
	// Returns the created URL model or an error if the operation fails.
	Create(ctx context.Context, shortCode, originalURL string) (*models.URL, error)

	// GetByShortCode retrieves a URL by its short code.
	// Returns the URL model if found or an error if not found.
	GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error)

	// Update modifies the original URL for a given short code.
	// Returns the updated URL model or an error if the operation fails.
	Update(ctx context.Context, shortCode, originalURL string) (*models.URL, error)

	// Delete removes a URL by its short code.
	// Returns an error if the operation fails.
	Delete(ctx context.Context, shortCode string) error

	// GetStats retrieves a URL by its short code without changing.
	// Returns the URL model if found or an error if not found.
	GetStats(ctx context.Context, shortCode string) (*models.URL, error)
}

// URLService provides methods to manage URL shortening operations.
// The service uses a URLRepository interface to interact with the underlying database.
type URLService struct {
	repo            URLRepository
	shortCodeLength int
}

// NewURLService creates a new instance of URLService with the provided repository and short code length.
func NewURLService(repo URLRepository, shortCodeLength int) *URLService {
	return &URLService{
		repo:            repo,
		shortCodeLength: shortCodeLength,
	}
}

// ShortenURL generates a short code for the provided original URL and stores it in the repository.
// It attempts to generate a unique short code up to a maximum number of retries.
// If successful, it returns the created URL model; otherwise, it returns an error.
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
				s.shortCodeLength++
				defer func() {
					s.shortCodeLength--
				}()

				continue
			}

			return nil, fmt.Errorf("%s: failed to shorten url: %w", op, err)
		}

		return url, nil
	}

	return nil, fmt.Errorf("%s: %w", op, ErrMaxRetriesExceeded)
}

// ResolveShortCode retrieves the original URL associated with the provided short code.
// If the short code exists, it returns the corresponding URL model; otherwise, it returns an error.
func (s *URLService) ResolveShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	const op = "service.URLService.ResolveShortCode"

	url, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to resolve short code: %w", op, err)
	}

	return url, nil
}

// ModifyURL updates the original URL associated with a given short code.
// It returns the updated URL model or an error if the operation fails.
func (s *URLService) ModifyURL(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	const op = "service.URLService.ModifyURL"

	url, err := s.repo.Update(ctx, shortCode, originalURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to modify url: %w", op, err)
	}

	return url, nil
}

// DeactivateURL deletes the URL associated with the provided short code.
// It returns an error if the deletion fails.
func (s *URLService) DeactivateURL(ctx context.Context, shortCode string) error {
	const op = "service.URLService.DeactivateURL"

	err := s.repo.Delete(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("%s: failed to deactivate url: %w", op, err)
	}

	return nil
}

// GetURLStats retrieves the statistics for the URL associated with the provided short code.
// It returns the URL model containing the statistics or an error if the operation fails.
func (s *URLService) GetURLStats(ctx context.Context, shortCode string) (*models.URL, error) {
	const op = "service.URLService.GetURLStats"

	url, err := s.repo.GetStats(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get url stats: %w", op, err)
	}

	return url, nil
}
