package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vadimbarashkov/url-shortener/internal/entity"
	"github.com/vadimbarashkov/url-shortener/mocks/usecase"
)

type URLUseCaseTestSuite struct {
	suite.Suite
	errUnknown  error
	urlRepoMock *usecase.MockUrlRepository
	uc          *URLUseCase
}

func (suite *URLUseCaseTestSuite) SetupSuite() {
	suite.errUnknown = errors.New("unknown error")
}

func (suite *URLUseCaseTestSuite) SetupSubTest() {
	suite.urlRepoMock = usecase.NewMockUrlRepository(suite.T())
	suite.uc = NewURLUseCase(suite.urlRepoMock)
}

func (suite *URLUseCaseTestSuite) TearDownSubTest() {
	suite.urlRepoMock.AssertExpectations(suite.T())
}

func (suite *URLUseCaseTestSuite) TestShortenURL() {
	suite.Run("short code generation error", func() {
		suite.uc.shortCodeLength = -1

		url, err := suite.uc.ShortenURL(context.Background(), "https://example.com")

		suite.Error(err)
		suite.Nil(url)
	})

	suite.Run("maximum retries error", func() {
		suite.urlRepoMock.
			On("Save", context.Background(), mock.Anything, "https://example.com").
			Times(5).
			Return(nil, entity.ErrShortCodeExists)

		url, err := suite.uc.ShortenURL(context.Background(), "https://example.com")

		suite.Error(err)
		suite.ErrorIs(err, ErrMaxRetriesExceeded)
		suite.Nil(url)
	})

	suite.Run("unknown error", func() {
		suite.urlRepoMock.
			On("Save", context.Background(), mock.Anything, "https://example.com").
			Once().
			Return(nil, suite.errUnknown)

		url, err := suite.uc.ShortenURL(context.Background(), "https://example.com")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		suite.urlRepoMock.
			On("Save", context.Background(), mock.Anything, "https://example.com").
			Once().
			Return(&entity.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://example.com",
				URLStats: entity.URLStats{
					AccessCount: 0,
				},
			}, nil)

		url, err := suite.uc.ShortenURL(context.Background(), "https://example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal(mock.Anything, url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Zero(url.URLStats.AccessCount)
	})
}

func (suite *URLUseCaseTestSuite) TestResolveShortCode() {
	suite.Run("unknown error", func() {
		suite.urlRepoMock.
			On("RetrieveAndUpdateStats", context.Background(), "abc123").
			Once().
			Return(nil, suite.errUnknown)

		url, err := suite.uc.ResolveShortCode(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		suite.urlRepoMock.
			On("RetrieveAndUpdateStats", context.Background(), "abc123").
			Once().
			Return(&entity.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				URLStats: entity.URLStats{
					AccessCount: 1,
				},
			}, nil)

		url, err := suite.uc.ResolveShortCode(context.Background(), "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Equal(int64(1), url.AccessCount)
	})
}

func (suite *URLUseCaseTestSuite) TestModifyURL() {
	suite.Run("unknown error", func() {
		suite.urlRepoMock.
			On("Update", context.Background(), "abc123", "https://new-example.com").
			Once().
			Return(nil, suite.errUnknown)

		url, err := suite.uc.ModifyURL(context.Background(), "abc123", "https://new-example.com")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		suite.urlRepoMock.
			On("Update", context.Background(), "abc123", "https://new-example.com").
			Once().
			Return(&entity.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				URLStats: entity.URLStats{
					AccessCount: 0,
				},
			}, nil)

		url, err := suite.uc.ModifyURL(context.Background(), "abc123", "https://new-example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Zero(url.URLStats.AccessCount)
	})
}

func (suite *URLUseCaseTestSuite) TestDeactivateURL() {
	suite.Run("unknown error", func() {
		suite.urlRepoMock.
			On("Remove", context.Background(), "abc123").
			Once().
			Return(suite.errUnknown)

		err := suite.uc.DeactivateURL(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
	})

	suite.Run("success", func() {
		suite.urlRepoMock.
			On("Remove", context.Background(), "abc123").
			Once().
			Return(nil)

		err := suite.uc.DeactivateURL(context.Background(), "abc123")

		suite.NoError(err)
	})
}

func (suite *URLUseCaseTestSuite) TestGetURLStats() {
	suite.Run("unknown error", func() {
		suite.urlRepoMock.
			On("RetrieveByShortCode", context.Background(), "abc123").
			Once().
			Return(nil, suite.errUnknown)

		url, err := suite.uc.GetURLStats(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		suite.urlRepoMock.
			On("RetrieveByShortCode", context.Background(), "abc123").
			Once().
			Return(&entity.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				URLStats: entity.URLStats{
					AccessCount: 1,
				},
			}, nil)

		url, err := suite.uc.GetURLStats(context.Background(), "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Equal(int64(1), url.AccessCount)
	})
}

func TestURLUseCase(t *testing.T) {
	suite.Run(t, new(URLUseCaseTestSuite))
}
