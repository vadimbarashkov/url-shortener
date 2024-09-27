// Package usecase implements the core business logic for managing URLs, including
// URL shortening, resolving short codes, modifying URLs, deactivating URLs, and
// retrieving URL statistics. It defines the URLUseCase structure, which encapsulates
// the logic for interacting with the URL repository and handling errors.
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/vadimbarashkov/url-shortener/internal/entity"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// ErrMaxRetriesExceeded is returned when the maximum number of retries for generating a unique short code is exceeded.
var ErrMaxRetriesExceeded = errors.New("maximum retries exceeded for generating short code")

// urlRepository defines the interface for interacting with the URL storage layer.
// Implementations of this interface must provide methods for saving, retrieving,
// updating, and removing URLs, as well as updating URL statistics.
type urlRepository interface {
	Save(ctx context.Context, shortCode, originalURL string) (*entity.URL, error)
	RetrieveByShortCode(ctx context.Context, shortCode string) (*entity.URL, error)
	RetrieveAndUpdateStats(ctx context.Context, shortCode string) (*entity.URL, error)
	Update(ctx context.Context, shortCode, originalURL string) (*entity.URL, error)
	Remove(ctx context.Context, shortCode string) error
}

// URLOption defines a functional option for configuring URLUseCase.
// It allows dynamic setting of use case parameters.
type URLOption func(*URLUseCase)

// WithMaxRetries sets the maximum number of retries for generating a unique short code.
func WithMaxRetries(n int) URLOption {
	return func(uc *URLUseCase) {
		uc.maxRetries = n
	}
}

// WithShortCodeLength sets the length of the generated short code for URL shortening.
func WithShortCodeLength(l int) URLOption {
	return func(uc *URLUseCase) {
		uc.shortCodeLength = l
	}
}

// URLUseCase is the main structure responsible for handling URL-related operations.
// It includes configuration for retries, short code length, and a reference to the repository for URL storage.
type URLUseCase struct {
	maxRetries      int
	shortCodeLength int
	urlRepo         urlRepository
}

// defaultURLUseCase provides default configuration values for URLUseCase.
var defaultURLUseCase = URLUseCase{
	maxRetries:      5,
	shortCodeLength: 7,
}

// NewURLUseCase creates a new instance of URLUseCase with the provided urlRepository and any functional options.
// It applies the default configuration and overrides them with provided options.
func NewURLUseCase(urlRepo urlRepository, opts ...URLOption) *URLUseCase {
	uc := defaultURLUseCase
	uc.urlRepo = urlRepo

	for _, opt := range opts {
		opt(&uc)
	}

	return &uc
}

// ShortenURL generates a unique short code for the provided original URL and saves it in the repository.
// It attempts to generate a unique short code, retrying up to maxRetries times if a conflict occurs.
func (uc *URLUseCase) ShortenURL(ctx context.Context, originalURL string) (*entity.URL, error) {
	const op = "usecase.URLUseCase.ShortenURL"

	shortCodeLength := uc.shortCodeLength

	for i := 0; i < uc.maxRetries; i++ {
		shortCode, err := gonanoid.New(shortCodeLength)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate short code: %w", op, err)
		}

		url, err := uc.urlRepo.Save(ctx, shortCode, originalURL)
		if err != nil {
			if errors.Is(err, entity.ErrShortCodeExists) {
				shortCodeLength++
				continue
			}

			return nil, fmt.Errorf("%s: failed to shorten url: %w", op, err)
		}

		return url, nil
	}

	return nil, fmt.Errorf("%s: %w", op, ErrMaxRetriesExceeded)
}

// ResolveShortCode retrieves the original URL corresponding to the provided short code,
// updating the access statistics in the process.
func (uc *URLUseCase) ResolveShortCode(ctx context.Context, shortCode string) (*entity.URL, error) {
	const op = "usecase.URLUseCase.ResolveShortCode"

	url, err := uc.urlRepo.RetrieveAndUpdateStats(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to resolve short code: %w", op, err)
	}

	return url, nil
}

// ModifyURL updates the original URL associated with the given short code in the repository.
func (uc *URLUseCase) ModifyURL(ctx context.Context, shortCode, originalURL string) (*entity.URL, error) {
	const op = "usecase.URLUseCase.ModifyURL"

	url, err := uc.urlRepo.Update(ctx, shortCode, originalURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to modify url: %w", op, err)
	}

	return url, nil
}

// DeactivateURL removes the URL associated with the given short code from the repository, effectively deactivating it.
func (uc *URLUseCase) DeactivateURL(ctx context.Context, shortCode string) error {
	const op = "usecase.URLUseCase.DeactivateURL"

	err := uc.urlRepo.Remove(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("%s: failed to deactivate url: %w", op, err)
	}

	return nil
}

// GetURLStats retrieves the URL associated with the given short code along with its usage statistics.
func (uc *URLUseCase) GetURLStats(ctx context.Context, shortCode string) (*entity.URL, error) {
	const op = "usecase.URLUseCase.GetURLStats"

	url, err := uc.urlRepo.RetrieveByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get url stats: %w", op, err)
	}

	return url, nil
}
