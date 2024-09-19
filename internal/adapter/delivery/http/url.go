package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/entity"
)

type urlUseCase interface {
	ShortenURL(ctx context.Context, originalURL string) (*entity.URL, error)
	ResolveShortCode(ctx context.Context, shortCode string) (*entity.URL, error)
	ModifyURL(ctx context.Context, shortCode, originalURL string) (*entity.URL, error)
	DeactivateURL(ctx context.Context, shortCode string) error
	GetURLStats(ctx context.Context, shortCode string) (*entity.URL, error)
}

type urlRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
}

type urlResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	AccessCount *int64    `json:"access_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type urlHandler struct {
	useCase  urlUseCase
	validate *validator.Validate
}

func newURLHandler(useCase urlUseCase, validate *validator.Validate) *urlHandler {
	return &urlHandler{
		useCase:  useCase,
		validate: validate,
	}
}

func (h *urlHandler) shortenURL(w http.ResponseWriter, r *http.Request) {
	var req urlRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"status":  "error",
				"message": "empty request body",
			})
			return
		}

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "invalid request body",
		})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "invalid values",
		})
		return
	}

	url, err := h.useCase.ShortenURL(r.Context(), req.OriginalURL)
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "server error",
		})
		return
	}

	resp := urlResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

func (h *urlHandler) resolveShortCode(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	url, err := h.useCase.ResolveShortCode(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "server error",
		})
		return
	}

	resp := urlResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

func (h *urlHandler) modifyURL(w http.ResponseWriter, r *http.Request) {
	var req urlRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{
				"status":  "error",
				"message": "empty request body",
			})
			return
		}

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "invalid request body",
		})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "invalid values",
		})
		return
	}

	shortCode := chi.URLParam(r, "shortCode")

	url, err := h.useCase.ModifyURL(r.Context(), shortCode, req.OriginalURL)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "server error",
		})
		return
	}

	resp := urlResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

func (h *urlHandler) deactivateURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	err := h.useCase.DeactivateURL(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "server error",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *urlHandler) getURLStats(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	url, err := h.useCase.GetURLStats(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"status":  "error",
			"message": "server error",
		})
		return
	}

	resp := urlResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		AccessCount: &url.AccessCount,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
