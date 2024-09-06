package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func setupRouter(t testing.TB) (*chi.Mux, *MockURLService) {
	t.Helper()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
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

		reqBody := bytes.NewBufferString(`{"url": "not url"}`)

		validate := getValidate()
		wantResp := response.ValidationErrorResponse(validate.Struct(urlRequest{"not url"}))

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
