package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func setupRouter(t testing.TB) (http.Handler, *MockURLService) {
	t.Helper()
	logger := httplog.NewLogger("", httplog.Options{Writer: io.Discard})
	mockURLSvc := new(MockURLService)
	return NewRouter(logger, mockURLSvc), mockURLSvc
}

func encode[T any](t testing.TB, v T) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestHandlePing(t *testing.T) {
	router, _ := setupRouter(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong\n", rec.Body.String())
}

func TestHandleShortenURL(t *testing.T) {
	t.Run("empty request body", func(t *testing.T) {
		router, _ := setupRouter(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.EmptyRequestBodyResponse), rec.Body.Bytes())
	})

	t.Run("invalid request body", func(t *testing.T) {
		router, _ := setupRouter(t)

		reqBody := bytes.NewBufferString(`invalid body`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.BadRequestResponse), rec.Body.Bytes())
	})

	t.Run("validation error", func(t *testing.T) {
		router, _ := setupRouter(t)

		validate := getValidate()
		wantResp := response.ValidationErrorResponse(validate.Struct(urlRequest{"not url"}))

		reqBody := bytes.NewBufferString(`{"url": "not url"}`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, wantResp), rec.Body.Bytes())
	})

	t.Run("server error", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ShortenURL", mock.Anything, "https://example.com").
			Times(1).
			Return(nil, errors.New("unknown error"))

		reqBody := bytes.NewBufferString(`{"url": "https://example.com"}`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ServerErrorResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ShortenURL", 1)
	})

	t.Run("success", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ShortenURL", mock.Anything, "https://example.com").
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://example.com",
			}, nil)

		wantResp := response.SuccessResponse("The URL has been shortened successfully.", urlResponse{
			ShortCode: mock.Anything,
			URL:       "https://example.com",
		})

		reqBody := bytes.NewBufferString(`{"url": "https://example.com"}`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, wantResp), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ShortenURL", 1)
	})
}

func TestHandleResolveShortCode(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ResolveShortCode", mock.Anything, mock.Anything).
			Times(1).
			Return(nil, database.ErrURLNotFound)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten/mock.Something", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ResourceNotFoundResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ResolveShortCode", 1)
	})

	t.Run("server error", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ResolveShortCode", mock.Anything, mock.Anything).
			Times(1).
			Return(nil, errors.New("unknown error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten/mock.Something", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ServerErrorResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ResolveShortCode", 1)
	})

	t.Run("success", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ResolveShortCode", mock.Anything, mock.Anything).
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://example.com",
			}, nil)

		wantResp := response.SuccessResponse("The short code was successfully resolved.", urlResponse{
			ShortCode: mock.Anything,
			URL:       "https://example.com",
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten/mock.Something", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, wantResp), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ResolveShortCode", 1)
	})
}

func TestHandleModifyURL(t *testing.T) {
	t.Run("empty request body", func(t *testing.T) {
		router, _ := setupRouter(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/shorten/mock.Anything", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.EmptyRequestBodyResponse), rec.Body.Bytes())
	})

	t.Run("invalid request body", func(t *testing.T) {
		router, _ := setupRouter(t)

		reqBody := bytes.NewBufferString(`invalid body`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/shorten/mock.Anything", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.BadRequestResponse), rec.Body.Bytes())
	})

	t.Run("validation error", func(t *testing.T) {
		router, _ := setupRouter(t)

		validate := getValidate()
		wantResp := response.ValidationErrorResponse(validate.Struct(urlRequest{"not url"}))

		reqBody := bytes.NewBufferString(`{"url": "not url"}`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/shorten/mock.Anything", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, wantResp), rec.Body.Bytes())
	})

	t.Run("not found", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ModifyURL", mock.Anything, mock.Anything, "https://new-example.com").
			Times(1).
			Return(nil, database.ErrURLNotFound)

		reqBody := bytes.NewBufferString(`{"url": "https://new-example.com"}`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/shorten/mock.Anything", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ResourceNotFoundResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ModifyURL", 1)
	})

	t.Run("server error", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ModifyURL", mock.Anything, mock.Anything, "https://new-example.com").
			Times(1).
			Return(nil, errors.New("unknown error"))

		reqBody := bytes.NewBufferString(`{"url": "https://new-example.com"}`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/shorten/mock.Anything", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ServerErrorResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ModifyURL", 1)
	})

	t.Run("success", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("ModifyURL", mock.Anything, mock.Anything, "https://new-example.com").
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://new-example.com",
			}, nil)

		wantResp := response.SuccessResponse("The URL was successfully modified.", urlResponse{
			ShortCode: mock.Anything,
			URL:       "https://new-example.com",
		})

		reqBody := bytes.NewBufferString(`{"url": "https://new-example.com"}`)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/shorten/mock.Anything", reqBody)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, wantResp), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "ModifyURL", 1)
	})
}

func TestHandleDeactivateURL(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("DeactivateURL", mock.Anything, mock.Anything).
			Times(1).
			Return(database.ErrURLNotFound)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/shorten/mock.Something", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ResourceNotFoundResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "DeactivateURL", 1)
	})

	t.Run("server error", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("DeactivateURL", mock.Anything, mock.Anything).
			Times(1).
			Return(errors.New("unknown error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/shorten/mock.Something", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ServerErrorResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "DeactivateURL", 1)
	})

	t.Run("success", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("DeactivateURL", mock.Anything, mock.Anything).
			Times(1).
			Return(nil)

		wantResp := response.SuccessResponse("The URL was successfully deactivated.")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/shorten/mock.Something", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, wantResp), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "DeactivateURL", 1)
	})
}

func TestHandleGetURLStats(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("GetURLStats", mock.Anything, mock.Anything).
			Times(1).
			Return(nil, database.ErrURLNotFound)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten/mock.Something/stats", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ResourceNotFoundResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "GetURLStats", 1)
	})

	t.Run("server error", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("GetURLStats", mock.Anything, mock.Anything).
			Times(1).
			Return(nil, errors.New("unknown error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten/mock.Something/stats", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, response.ServerErrorResponse), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "GetURLStats", 1)
	})

	t.Run("success", func(t *testing.T) {
		router, mockURLSvc := setupRouter(t)

		mockURLSvc.On("GetURLStats", mock.Anything, mock.Anything).
			Times(1).
			Return(&models.URL{
				ShortCode:   mock.Anything,
				OriginalURL: "https://example.com",
				AccessCount: 1,
			}, nil)

		wantResp := response.SuccessResponse("The URL statistics retrieved successfully.", urlResponse{
			ShortCode:   mock.Anything,
			URL:         "https://example.com",
			AccessCount: 1,
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten/mock.Something/stats", nil)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, encode(t, wantResp), rec.Body.Bytes())
		mockURLSvc.AssertExpectations(t)
		mockURLSvc.AssertNumberOfCalls(t, "GetURLStats", 1)
	})
}
