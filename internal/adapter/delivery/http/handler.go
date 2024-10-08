package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/entity"
)

// handlePing handles the ping request and responds with "pong".
// This is a simple health check endpoint.
func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
}

// urlUseCase defines the methods required for URL shortening and management.
// It abstracts the business logic needed for handling URLs.
type urlUseCase interface {
	ShortenURL(ctx context.Context, originalURL string) (*entity.URL, error)
	ResolveShortCode(ctx context.Context, shortCode string) (*entity.URL, error)
	ModifyURL(ctx context.Context, shortCode, originalURL string) (*entity.URL, error)
	DeactivateURL(ctx context.Context, shortCode string) error
	GetURLStats(ctx context.Context, shortCode string) (*entity.URL, error)
}

// urlHandler handles HTTP requests related to URLs.
type urlHandler struct {
	useCase  urlUseCase
	validate *validator.Validate
}

// newURLHandler creates a new instance of urlHandler with the provided use case and validator.
func newURLHandler(useCase urlUseCase, validate *validator.Validate) *urlHandler {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &urlHandler{
		useCase:  useCase,
		validate: validate,
	}
}

// shortenURL handles the request to shorten a URL.
func (h *urlHandler) shortenURL(w http.ResponseWriter, r *http.Request) {
	var req urlRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, emptyRequestBodyResponse)
			return
		}

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidRequestBodyResponse)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, validationErrorResponse(err))
		return
	}

	url, err := h.useCase.ShortenURL(r.Context(), req.OriginalURL)
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrorResponse)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, toURLResponse(url))
}

// resolveShortCode handles the request to resolve a shortened URL.
func (h *urlHandler) resolveShortCode(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	url, err := h.useCase.ResolveShortCode(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, urlNotFoundResponse)
			return
		}

		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrorResponse)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toURLResponse(url))
}

// modifyURL handles the request to modify an existing shortened URL.
func (h *urlHandler) modifyURL(w http.ResponseWriter, r *http.Request) {
	var req urlRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, emptyRequestBodyResponse)
			return
		}

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidRequestBodyResponse)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, validationErrorResponse(err))
		return
	}

	shortCode := chi.URLParam(r, "shortCode")

	url, err := h.useCase.ModifyURL(r.Context(), shortCode, req.OriginalURL)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, urlNotFoundResponse)
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrorResponse)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toURLResponse(url))
}

// deactivateURL handles the request to deactivate a shortened URL.
func (h *urlHandler) deactivateURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	err := h.useCase.DeactivateURL(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, urlNotFoundResponse)
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrorResponse)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getURLStats handles the request to retrieve statistics for a shortened URL.
func (h *urlHandler) getURLStats(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	url, err := h.useCase.GetURLStats(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, urlNotFoundResponse)
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrorResponse)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toURLStatsResponse(url))
}
