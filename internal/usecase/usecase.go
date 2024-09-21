package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/vadimbarashkov/url-shortener/internal/entity"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ErrMaxRetriesExceeded = errors.New("maximum retries exceeded for generating short code")

type urlRepository interface {
	Save(ctx context.Context, shortCode, originalURL string) (*entity.URL, error)
	RetrieveByShortCode(ctx context.Context, shortCode string) (*entity.URL, error)
	RetrieveAndUpdateStats(ctx context.Context, shortCode string) (*entity.URL, error)
	Update(ctx context.Context, shortCode, originalURL string) (*entity.URL, error)
	Remove(ctx context.Context, shortCode string) error
}

type URLOption func(*URLUseCase)

func WithMaxRetries(n int) URLOption {
	return func(uc *URLUseCase) {
		uc.maxRetries = n
	}
}

func WithShortCodeLength(l int) URLOption {
	return func(uc *URLUseCase) {
		uc.shortCodeLength = l
	}
}

type URLUseCase struct {
	maxRetries      int
	shortCodeLength int
	urlRepo         urlRepository
}

var defaultURLUseCase = URLUseCase{
	maxRetries:      5,
	shortCodeLength: 7,
}

func NewURLUseCase(urlRepo urlRepository, opts ...URLOption) *URLUseCase {
	uc := defaultURLUseCase
	uc.urlRepo = urlRepo

	for _, opt := range opts {
		opt(&uc)
	}

	return &uc
}

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

func (uc *URLUseCase) ResolveShortCode(ctx context.Context, shortCode string) (*entity.URL, error) {
	const op = "usecase.URLUseCase.ResolveShortCode"

	url, err := uc.urlRepo.RetrieveAndUpdateStats(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to resolve short code: %w", op, err)
	}

	return url, nil
}

func (uc *URLUseCase) ModifyURL(ctx context.Context, shortCode, originalURL string) (*entity.URL, error) {
	const op = "usecase.URLUseCase.ModifyURL"

	url, err := uc.urlRepo.Update(ctx, shortCode, originalURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to modify url: %w", op, err)
	}

	return url, nil
}

func (uc *URLUseCase) DeactivateURL(ctx context.Context, shortCode string) error {
	const op = "usecase.URLUseCase.DeactivateURL"

	err := uc.urlRepo.Remove(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("%s: failed to deactivate url: %w", op, err)
	}

	return nil
}

func (uc *URLUseCase) GetURLStats(ctx context.Context, shortCode string) (*entity.URL, error) {
	const op = "usecase.URLUseCase.GetURLStats"

	url, err := uc.urlRepo.RetrieveByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get url stats: %w", op, err)
	}

	return url, nil
}
