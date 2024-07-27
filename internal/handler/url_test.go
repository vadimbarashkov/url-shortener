package handler_test

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vadimbarashkov/url-shortener/internal/handler"
	"github.com/vadimbarashkov/url-shortener/internal/storage"
	mock_storage "github.com/vadimbarashkov/url-shortener/internal/storage/mock"
)

var ErrUnknown = errors.New("unknown error")

func TestURLHandler_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlStorage := mock_storage.NewMockURLStorage(ctrl)
	handler := handler.NewURLHandler(urlStorage)

	app := fiber.New()
	app.Post("/urls", handler.Add)

	t.Run("status bad request after request body parse", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/urls", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("status bad request after request body validation", func(t *testing.T) {
		reqBody := strings.NewReader(`{"alias": "alias", "url": "invalid url"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/urls", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("status conflict", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Add(gomock.Any(), "alias", "https://example.com").
			Return(storage.ErrURLExists)

		reqBody := strings.NewReader(`{"alias": "alias", "url": "https://example.com"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/urls", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	})

	t.Run("status internal server error", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Add(gomock.Any(), "alias", "https://example.com").
			Return(ErrUnknown)

		reqBody := strings.NewReader(`{"alias": "alias", "url": "https://example.com"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/urls", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("status created", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Add(gomock.Any(), "alias", "https://example.com").
			Return(nil)

		reqBody := strings.NewReader(`{"alias": "alias", "url": "https://example.com"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/urls", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})
}

func TestURLHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlStorage := mock_storage.NewMockURLStorage(ctrl)
	handler := handler.NewURLHandler(urlStorage)

	app := fiber.New()
	app.Get("/urls/:alias", handler.Get)

	t.Run("status not found", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Get(gomock.Any(), "alias").
			Return("", storage.ErrURLNotFound)

		req := httptest.NewRequest(fiber.MethodGet, "/urls/alias", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("internal server error", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Get(gomock.Any(), "alias").
			Return("", ErrUnknown)

		req := httptest.NewRequest(fiber.MethodGet, "/urls/alias", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("status found", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Get(gomock.Any(), "alias").
			Return("https://example.com", nil)

		req := httptest.NewRequest(fiber.MethodGet, "/urls/alias", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusFound, resp.StatusCode)
	})
}

func TestURLHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlStorage := mock_storage.NewMockURLStorage(ctrl)
	handler := handler.NewURLHandler(urlStorage)

	app := fiber.New()
	app.Put("/urls/:alias", handler.Update)

	t.Run("status bad request after request body parse", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/urls/alias", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("status bad request after request body validation", func(t *testing.T) {
		reqBody := strings.NewReader(`{"url": "invalid url"}`)
		req := httptest.NewRequest(fiber.MethodPut, "/urls/alias", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("status not found", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Update(gomock.Any(), "alias", "https://example.com").
			Return(storage.ErrURLNotFound)

		reqBody := strings.NewReader(`{"url": "https://example.com"}`)
		req := httptest.NewRequest(fiber.MethodPut, "/urls/alias", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("status internal server error", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Update(gomock.Any(), "alias", "https://example.com").
			Return(ErrUnknown)

		reqBody := strings.NewReader(`{"url": "https://example.com"}`)
		req := httptest.NewRequest(fiber.MethodPut, "/urls/alias", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("status no content", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Update(gomock.Any(), "alias", "https://example.com").
			Return(nil)

		reqBody := strings.NewReader(`{"url": "https://example.com"}`)
		req := httptest.NewRequest(fiber.MethodPut, "/urls/alias", reqBody)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	})
}

func TestURLHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlStorage := mock_storage.NewMockURLStorage(ctrl)
	handler := handler.NewURLHandler(urlStorage)

	app := fiber.New()
	app.Delete("/urls/:alias", handler.Delete)

	t.Run("status not found", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Delete(gomock.Any(), "alias").
			Return(storage.ErrURLNotFound)

		req := httptest.NewRequest(fiber.MethodDelete, "/urls/alias", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("status internal server error", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Delete(gomock.Any(), "alias").
			Return(ErrUnknown)

		req := httptest.NewRequest(fiber.MethodDelete, "/urls/alias", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("status no content", func(t *testing.T) {
		urlStorage.
			EXPECT().
			Delete(gomock.Any(), "alias").
			Return(nil)

		req := httptest.NewRequest(fiber.MethodDelete, "/urls/alias", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	})
}
