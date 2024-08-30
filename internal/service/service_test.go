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

		mockRepo.On("Create", context.TODO(), mock.Anything, "https://example.com").
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

		mockRepo.On("Create", context.TODO(), mock.Anything, "https://example.com").
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

		mockRepo.On("Create", context.TODO(), mock.Anything, "https://example.com").
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

func TestURLService_ResolveShortCode(t *testing.T) {
	t.Run("unknown error", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("GetByShortCode", context.TODO(), mock.Anything).
			Times(1).
			Return(nil, errUnknown)

		url, err := svc.ResolveShortCode(context.TODO(), mock.Anything)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, url)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "GetByShortCode", 1)
	})

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("GetByShortCode", context.TODO(), mock.Anything).
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://example.com",
				AccessCount: 1,
			}, nil)

		wantURL := models.URL{
			ShortCode:   mock.Anything,
			OriginalURL: "https://example.com",
			AccessCount: 1,
		}

		url, err := svc.ResolveShortCode(context.TODO(), mock.Anything)

		assert.NoError(t, err)
		assert.NotNil(t, url)
		assert.Equal(t, wantURL, *url)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "GetByShortCode", 1)
	})
}

func TestURLService_ModifyURL(t *testing.T) {
	t.Run("unknown error", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("Update", context.TODO(), mock.Anything, "https://new-example.com").
			Times(1).
			Return(nil, errUnknown)

		url, err := svc.ModifyURL(context.TODO(), mock.Anything, "https://new-example.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, url)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("Update", context.TODO(), mock.Anything, "https://new-example.com").
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://new-example.com",
			}, nil)

		wantURL := models.URL{
			ShortCode:   mock.Anything,
			OriginalURL: "https://new-example.com",
		}

		url, err := svc.ModifyURL(context.TODO(), mock.Anything, "https://new-example.com")

		assert.NoError(t, err)
		assert.NotNil(t, url)
		assert.Equal(t, wantURL, *url)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "Update", 1)
	})
}

func TestURLService_DeactivateURL(t *testing.T) {
	t.Run("unknown error", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("Delete", context.TODO(), mock.Anything).
			Times(1).
			Return(errUnknown)

		err := svc.DeactivateURL(context.TODO(), mock.Anything)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "Delete", 1)
	})

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := setupURLSerive(t, 8)

		mockRepo.On("Delete", context.TODO(), mock.Anything).
			Times(1).
			Return(nil)

		err := svc.DeactivateURL(context.TODO(), mock.Anything)

		assert.Nil(t, err)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNumberOfCalls(t, "Delete", 1)
	})
}
