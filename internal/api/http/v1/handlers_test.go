package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
	"github.com/vadimbarashkov/url-shortener/pkg/response"
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

type HandlersTestSuite struct {
	suite.Suite
	logger     *httplog.Logger
	urlSvcMock *MockURLService
	server     *httptest.Server
	e          *httpexpect.Expect
}

func (suite *HandlersTestSuite) SetupSuite() {
	suite.logger = httplog.NewLogger("", httplog.Options{Writer: io.Discard})
}

func (suite *HandlersTestSuite) SetupSubTest() {
	suite.urlSvcMock = new(MockURLService)
	router := NewRouter(suite.logger, suite.urlSvcMock)
	suite.server = httptest.NewServer(router)
	suite.e = httpexpect.Default(suite.T(), suite.server.URL)
}

func (suite *HandlersTestSuite) TeadDownSubTest() {
	suite.urlSvcMock.AssertExpectations(suite.T())
	suite.server.Close()
}

func (suite *HandlersTestSuite) TestPing() {
	const path = "/api/v1/ping"

	suite.Run("success", func() {
		suite.e.GET(path).
			Expect().
			Status(http.StatusOK).
			Text().IsEqual("pong\n")
	})
}

func (suite *HandlersTestSuite) TestShortenURL() {
	const path = "/api/v1/shorten"

	suite.Run("empty request body", func() {
		suite.e.POST(path).
			Expect().
			Status(http.StatusBadRequest).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.EmptyRequestBodyResponse.Message)
	})

	suite.Run("invalid request body", func() {
		suite.e.POST(path).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.BadRequestResponse.Message)
	})

	suite.Run("validation error", func() {
		suite.e.POST(path).
			WithJSON(map[string]string{
				"url": "invalid url",
			}).
			Expect().
			Status(http.StatusBadRequest).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			ContainsKey("message").
			ContainsKey("details")
	})

	suite.Run("server error", func() {
		suite.urlSvcMock.
			On("ShortenURL", mock.Anything, "https://example.com").
			Times(1).
			Return(nil, errors.New("unknown error"))

		suite.e.POST(path).
			WithJSON(map[string]string{
				"url": "https://example.com",
			}).
			Expect().
			Status(http.StatusInternalServerError).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ServerErrorResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ShortenURL", 1)
	})

	suite.Run("success", func() {
		suite.urlSvcMock.
			On("ShortenURL", mock.Anything, "https://example.com").
			Times(1).
			Return(&models.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
			}, nil)

		suite.e.POST(path).
			WithJSON(map[string]string{
				"url": "https://example.com",
			}).
			Expect().
			Status(http.StatusCreated).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusSuccess).
			ContainsKey("message").
			Value("data").Object().
			HasValue("short_code", "abc123").
			HasValue("url", "https://example.com")

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ShortenURL", 1)
	})
}

func (suite *HandlersTestSuite) TestResolveShortCode() {
	const path = "/api/v1/shorten/%s"

	suite.Run("not found", func() {
		suite.urlSvcMock.
			On("ResolveShortCode", mock.Anything, "abc123").
			Times(1).
			Return(nil, database.ErrURLNotFound)

		suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ResourceNotFoundResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ResolveShortCode", 1)
	})

	suite.Run("server error", func() {
		suite.urlSvcMock.
			On("ResolveShortCode", mock.Anything, "abc123").
			Times(1).
			Return(nil, errors.New("unknown error"))

		suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusInternalServerError).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ServerErrorResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ResolveShortCode", 1)
	})

	suite.Run("success", func() {
		suite.urlSvcMock.
			On("ResolveShortCode", mock.Anything, "abc123").
			Times(1).
			Return(&models.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				AccessCount: 1,
			}, nil)

		suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusOK).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusSuccess).
			ContainsKey("message").
			Value("data").Object().
			HasValue("short_code", "abc123").
			HasValue("url", "https://example.com").
			NotContainsKey("access_count")

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ResolveShortCode", 1)
	})
}

func (suite *HandlersTestSuite) TestModifyURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("empty request body", func() {
		suite.e.PUT(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusBadRequest).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.EmptyRequestBodyResponse.Message)
	})

	suite.Run("invalid request body", func() {
		suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.BadRequestResponse.Message)
	})

	suite.Run("validation error", func() {
		suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{
				"url": "invalid url",
			}).
			Expect().
			Status(http.StatusBadRequest).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			ContainsKey("message").
			ContainsKey("details")
	})

	suite.Run("not found", func() {
		suite.urlSvcMock.
			On("ModifyURL", mock.Anything, "abc123", "https://new-example.com").
			Times(1).
			Return(nil, database.ErrURLNotFound)

		suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{
				"url": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusNotFound).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ResourceNotFoundResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ModifyURL", 1)
	})

	suite.Run("server error", func() {
		suite.urlSvcMock.
			On("ModifyURL", mock.Anything, "abc123", "https://new-example.com").
			Times(1).
			Return(nil, errors.New("unknown error"))

		suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{
				"url": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusInternalServerError).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ServerErrorResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ModifyURL", 1)
	})

	suite.Run("success", func() {
		suite.urlSvcMock.
			On("ModifyURL", mock.Anything, "abc123", "https://new-example.com").
			Times(1).
			Return(&models.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://new-example.com",
			}, nil)

		suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{
				"url": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusOK).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusSuccess).
			ContainsKey("message").
			Value("data").Object().
			HasValue("short_code", "abc123").
			HasValue("url", "https://new-example.com")

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "ModifyURL", 1)
	})
}

func (suite *HandlersTestSuite) TestDeactivateURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("not found", func() {
		suite.urlSvcMock.
			On("DeactivateURL", mock.Anything, "abc123").
			Times(1).
			Return(database.ErrURLNotFound)

		suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ResourceNotFoundResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "DeactivateURL", 1)
	})

	suite.Run("server error", func() {
		suite.urlSvcMock.
			On("DeactivateURL", mock.Anything, "abc123").
			Times(1).
			Return(errors.New("unknown error"))

		suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusInternalServerError).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ServerErrorResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "DeactivateURL", 1)
	})

	suite.Run("success", func() {
		suite.urlSvcMock.
			On("DeactivateURL", mock.Anything, "abc123").
			Times(1).
			Return(nil)

		suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusOK).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusSuccess).
			ContainsKey("message")

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "DeactivateURL", 1)
	})
}

func (suite *HandlersTestSuite) TestGetURLStats() {
	const path = "/api/v1/shorten/%s/stats"

	suite.Run("not found", func() {
		suite.urlSvcMock.
			On("GetURLStats", mock.Anything, "abc123").
			Times(1).
			Return(nil, database.ErrURLNotFound)

		suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ResourceNotFoundResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "GetURLStats", 1)
	})

	suite.Run("server error", func() {
		suite.urlSvcMock.
			On("GetURLStats", mock.Anything, "abc123").
			Times(1).
			Return(nil, errors.New("unknown error"))

		suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusInternalServerError).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusError).
			HasValue("message", response.ServerErrorResponse.Message)

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "GetURLStats", 1)
	})

	suite.Run("success", func() {
		suite.urlSvcMock.
			On("GetURLStats", mock.Anything, "abc123").
			Times(1).
			Return(&models.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				AccessCount: 1,
			}, nil)

		suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusOK).
			HasContentType("application/json").
			JSON().Object().
			HasValue("status", response.StatusSuccess).
			ContainsKey("message").
			Value("data").Object().
			HasValue("short_code", "abc123").
			HasValue("url", "https://example.com").
			HasValue("access_count", int64(1))

		suite.urlSvcMock.AssertNumberOfCalls(suite.T(), "GetURLStats", 1)
	})
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
