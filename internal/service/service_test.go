package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

var errUnknown = errors.New("unknown error")

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

func setupURLSerive(t testing.TB, shortCodeLength int) (*URLService, *MockURLRepository) {
	t.Helper()

	mockRepo := new(MockURLRepository)
	svc := NewURLService(mockRepo, shortCodeLength)

	return svc, mockRepo
}

func TestURLService_ShortenURL(t *testing.T) {
	t.Run("short code generation error", func(t *testing.T) {
		svc, _ := setupURLSerive(t, -1)

		url, err := svc.ShortenURL(context.TODO(), "https://example.com")

		assert.Error(t, err)
		assert.Nil(t, url)
	})

	t.Run("maximum retries error", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).
			Times(5).
			Return(nil, database.ErrShortCodeExists)

		url, err := svc.ShortenURL(context.TODO(), "https://example.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMaxRetriesExceeded)
		assert.Nil(t, url)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "Create", 5)
	})

	t.Run("unknown error", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).
			Times(1).
			Return(nil, errUnknown)

		url, err := svc.ShortenURL(context.TODO(), "https://example.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, url)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "Create", 1)
	})

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://example.com",
			}, nil)

		wantURL := models.URL{
			ShortCode:   mock.Anything,
			OriginalURL: "https://example.com",
		}

		url, err := svc.ShortenURL(context.TODO(), "https://example.com")

		assert.NoError(t, err)
		assert.NotNil(t, url)
		assert.Equal(t, wantURL, *url)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "Create", 1)
	})
}
