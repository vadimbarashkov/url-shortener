package http

import (
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

	"github.com/vadimbarashkov/url-shortener/internal/entity"

	httpMock "github.com/vadimbarashkov/url-shortener/mocks/http"
)

type HandlersTestSuite struct {
	suite.Suite
	logger         *httplog.Logger
	urlUseCaseMock *httpMock.MockUrlUseCase
	server         *httptest.Server
	e              *httpexpect.Expect
}

func (suite *HandlersTestSuite) SetupSuite() {
	suite.logger = httplog.NewLogger("", httplog.Options{Writer: io.Discard})
}

func (suite *HandlersTestSuite) SetupSubTest() {
	suite.urlUseCaseMock = httpMock.NewMockUrlUseCase(suite.T())

	router := NewRouter(suite.logger, suite.urlUseCaseMock)
	suite.server = httptest.NewServer(router)
	suite.T().Cleanup(func() {
		suite.server.Close()
	})

	suite.e = httpexpect.Default(suite.T(), suite.server.URL)
}

func (suite *HandlersTestSuite) TearDownSubTest() {
	suite.urlUseCaseMock.AssertExpectations(suite.T())
}

func (suite *HandlersTestSuite) TestPing() {
	const path = "/api/v1/ping"

	suite.Run("success", func() {
		suite.e.GET(path).
			Expect().
			Status(http.StatusOK).
			Text().IsEqual("pong")
	})
}

func (suite *HandlersTestSuite) TestShortenURL() {
	const path = "/api/v1/shorten"

	suite.Run("empty request body", func() {
		resp := suite.e.POST(path).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("invalid request body", func() {
		resp := suite.e.POST(path).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("validation error", func() {
		resp := suite.e.POST(path).
			WithJSON(map[string]string{"original_url": "invalid url"}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
		resp.Value("errors").Array().Value(0).Object().
			HasValue("field", "original_url").
			ContainsKey("message")
	})

	suite.Run("server error", func() {
		suite.urlUseCaseMock.
			On("ShortenURL", mock.Anything, "https://example.com").
			Once().
			Return(nil, errors.New("unknown error"))

		resp := suite.e.POST(path).
			WithJSON(map[string]string{"original_url": "https://example.com"}).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		suite.urlUseCaseMock.
			On("ShortenURL", mock.Anything, "https://example.com").
			Once().
			Return(&entity.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
			}, nil)

		resp := suite.e.POST(path).
			WithJSON(map[string]string{"original_url": "https://example.com"}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.ContainsKey("id")
		resp.HasValue("short_code", "abc123")
		resp.HasValue("original_url", "https://example.com")
		resp.NotContainsKey("stats")
		resp.ContainsKey("created_at")
		resp.ContainsKey("updated_at")
	})
}

func (suite *HandlersTestSuite) TestResolveShortCode() {
	path := "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		suite.urlUseCaseMock.
			On("ResolveShortCode", mock.Anything, "abc123").
			Once().
			Return(nil, entity.ErrURLNotFound)

		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("server error", func() {
		suite.urlUseCaseMock.
			On("ResolveShortCode", mock.Anything, "abc123").
			Once().
			Return(nil, errors.New("unknown error"))

		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		suite.urlUseCaseMock.
			On("ResolveShortCode", mock.Anything, "abc123").
			Once().
			Return(&entity.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				URLStats: entity.URLStats{
					AccessCount: 1,
				},
			}, nil)

		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.ContainsKey("id")
		resp.HasValue("short_code", "abc123")
		resp.HasValue("original_url", "https://example.com")
		resp.NotContainsKey("stats")
		resp.ContainsKey("created_at")
		resp.ContainsKey("updated_at")
	})
}

func (suite *HandlersTestSuite) TestModifyURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("empty request body", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("invalid request body", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("validation error", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{"original_url": "invalid url"}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
		resp.Value("errors").Array().Value(0).Object().
			HasValue("field", "original_url").
			ContainsKey("message")
	})

	suite.Run("url not found", func() {
		suite.urlUseCaseMock.
			On("ModifyURL", mock.Anything, "abc123", "https://new-example.com").
			Once().
			Return(nil, entity.ErrURLNotFound)

		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{"original_url": "https://new-example.com"}).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("server error", func() {
		suite.urlUseCaseMock.
			On("ModifyURL", mock.Anything, "abc123", "https://new-example.com").
			Once().
			Return(nil, errors.New("unknown error"))

		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{"original_url": "https://new-example.com"}).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		suite.urlUseCaseMock.
			On("ModifyURL", mock.Anything, "abc123", "https://new-example.com").
			Once().
			Return(&entity.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://new-example.com",
			}, nil)

		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{"original_url": "https://new-example.com"}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.ContainsKey("id")
		resp.HasValue("short_code", "abc123")
		resp.HasValue("original_url", "https://new-example.com")
		resp.NotContainsKey("stats")
		resp.ContainsKey("created_at")
		resp.ContainsKey("updated_at")
	})
}

func (suite *HandlersTestSuite) TestDeactivateURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		suite.urlUseCaseMock.
			On("DeactivateURL", mock.Anything, "abc123").
			Once().
			Return(entity.ErrURLNotFound)

		resp := suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("server error", func() {
		suite.urlUseCaseMock.
			On("DeactivateURL", mock.Anything, "abc123").
			Once().
			Return(errors.New("unknown error"))

		resp := suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		suite.urlUseCaseMock.
			On("DeactivateURL", mock.Anything, "abc123").
			Once().
			Return(nil)

		suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNoContent)
	})
}

func (suite *HandlersTestSuite) TestGetURLStats() {
	path := "/api/v1/shorten/%s/stats"

	suite.Run("url not found", func() {
		suite.urlUseCaseMock.
			On("GetURLStats", mock.Anything, "abc123").
			Once().
			Return(nil, entity.ErrURLNotFound)

		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("server error", func() {
		suite.urlUseCaseMock.
			On("GetURLStats", mock.Anything, "abc123").
			Once().
			Return(nil, errors.New("unknown error"))

		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		suite.urlUseCaseMock.
			On("GetURLStats", mock.Anything, "abc123").
			Once().
			Return(&entity.URL{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				URLStats: entity.URLStats{
					AccessCount: 1,
				},
			}, nil)

		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.ContainsKey("id")
		resp.HasValue("short_code", "abc123")
		resp.HasValue("original_url", "https://example.com")
		resp.Value("stats").Object().
			HasValue("access_count", int64(1))
		resp.ContainsKey("created_at")
		resp.ContainsKey("updated_at")
	})
}

func TestURLHandler(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
