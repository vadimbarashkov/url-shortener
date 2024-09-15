package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vadimbarashkov/url-shortener/internal/database"
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

type URLServiceTestSuite struct {
	suite.Suite
	mock       *MockURLRepository
	svc        *URLService
	errUnknown error
}

func (suite *URLServiceTestSuite) SetupSuite() {
	suite.errUnknown = errors.New("unknown error")
}

func (suite *URLServiceTestSuite) SetupSubTest() {
	suite.mock = new(MockURLRepository)
	suite.svc = NewURLService(suite.mock, 7)
}

func (suite *URLServiceTestSuite) TearDownSubTest() {
	suite.mock.AssertExpectations(suite.T())
}

func (suite *URLServiceTestSuite) TestShortenURL() {
	suite.Run("short code generation error", func() {
		suite.svc.shortCodeLength = -1

		url, err := suite.svc.ShortenURL(context.Background(), "https://example.com")

		suite.Error(err)
		suite.Nil(url)
	})

	suite.Run("maximum retries error", func() {
		suite.mock.
			On("Create", context.Background(), mock.Anything, "https://example.com").
			Times(5).
			Return(nil, database.ErrShortCodeExists)

		url, err := suite.svc.ShortenURL(context.Background(), "https://example.com")

		suite.Error(err)
		suite.ErrorIs(err, ErrMaxRetriesExceeded)
		suite.Nil(url)
		suite.mock.AssertNumberOfCalls(suite.T(), "Create", 5)
	})

	suite.Run("unknown error", func() {
		suite.mock.
			On("Create", context.Background(), mock.Anything, "https://example.com").
			Times(1).
			Return(nil, suite.errUnknown)

		url, err := suite.svc.ShortenURL(context.Background(), "https://example.com")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
		suite.mock.AssertNumberOfCalls(suite.T(), "Create", 1)
	})

	suite.Run("success", func() {
		suite.mock.
			On("Create", context.Background(), mock.Anything, "https://example.com").
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://example.com",
			}, nil)

		url, err := suite.svc.ShortenURL(context.Background(), "https://example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal(mock.Anything, url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.mock.AssertNumberOfCalls(suite.T(), "Create", 1)
	})
}

func (suite *URLServiceTestSuite) TestResolveShortCode() {
	suite.Run("unknown error", func() {
		suite.mock.
			On("GetByShortCode", context.Background(), "abc123").
			Times(1).
			Return(nil, suite.errUnknown)

		url, err := suite.svc.ResolveShortCode(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
		suite.mock.AssertNumberOfCalls(suite.T(), "GetByShortCode", 1)
	})

	suite.Run("success", func() {
		suite.mock.
			On("GetByShortCode", context.Background(), "abc123").
			Times(1).
			Return(&models.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				AccessCount: 1,
			}, nil)

		url, err := suite.svc.ResolveShortCode(context.Background(), "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Equal(int64(1), url.AccessCount)
		suite.mock.AssertNumberOfCalls(suite.T(), "GetByShortCode", 1)
	})
}

func (suite *URLServiceTestSuite) TestModifyURL() {
	suite.Run("unknown error", func() {
		suite.mock.
			On("Update", context.Background(), "abc123", "https://new-example.com").
			Times(1).
			Return(nil, suite.errUnknown)

		url, err := suite.svc.ModifyURL(context.Background(), "abc123", "https://new-example.com")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
		suite.mock.AssertNumberOfCalls(suite.T(), "Update", 1)
	})

	suite.Run("success", func() {
		suite.mock.
			On("Update", context.Background(), "abc123", "https://new-example.com").
			Times(1).
			Return(&models.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
			}, nil)

		url, err := suite.svc.ModifyURL(context.Background(), "abc123", "https://new-example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.mock.AssertNumberOfCalls(suite.T(), "Update", 1)
	})
}

func (suite *URLServiceTestSuite) TestDeactivateURLL() {
	suite.Run("unknown error", func() {
		suite.mock.
			On("Delete", context.Background(), "abc123").
			Times(1).
			Return(suite.errUnknown)

		err := suite.svc.DeactivateURL(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.mock.AssertNumberOfCalls(suite.T(), "Delete", 1)
	})

	suite.Run("success", func() {
		suite.mock.
			On("Delete", context.Background(), "abc123").
			Times(1).
			Return(nil)

		err := suite.svc.DeactivateURL(context.Background(), "abc123")

		suite.NoError(err)
		suite.mock.AssertNumberOfCalls(suite.T(), "Delete", 1)
	})
}

func (suite *URLServiceTestSuite) TestGetURLStats() {
	suite.Run("unknown error", func() {
		suite.mock.
			On("GetStats", context.Background(), "abc123").
			Times(1).
			Return(nil, suite.errUnknown)

		url, err := suite.svc.GetURLStats(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
		suite.mock.AssertNumberOfCalls(suite.T(), "GetStats", 1)
	})

	suite.Run("success", func() {
		suite.mock.
			On("GetStats", context.Background(), "abc123").
			Times(1).
			Return(&models.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				AccessCount: 1,
			}, nil)

		url, err := suite.svc.GetURLStats(context.Background(), "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Equal(int64(1), url.AccessCount)
		suite.mock.AssertNumberOfCalls(suite.T(), "GetStats", 1)
	})
}

func TestURLService(t *testing.T) {
	suite.Run(t, new(URLServiceTestSuite))
}
