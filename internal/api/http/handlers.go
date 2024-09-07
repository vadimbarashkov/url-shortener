package http

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
	"github.com/vadimbarashkov/url-shortener/pkg/response"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
}

type urlRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type urlResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	URL         string    `json:"url"`
	AccessCount int64     `json:"access_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toURLResponse(url *models.URL) urlResponse {
	return urlResponse{
		ID:        url.ID,
		ShortCode: url.ShortCode,
		URL:       url.OriginalURL,
		CreatedAt: url.CreatedAt,
		UpdatedAt: url.UpdatedAt,
	}
}

func handleShortenURL(logger *slog.Logger, svc URLService, validate *validator.Validate) http.HandlerFunc {
	const op = "api.http.handleShortenURL"
	const successMsg = "The URL has been shortened successfully."

	return func(w http.ResponseWriter, r *http.Request) {
		var req urlRequest

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			if errors.Is(err, io.EOF) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.EmptyRequestBodyResponse)
				return
			}

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.BadRequestResponse)
			return
		}

		if err := validate.Struct(req); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationErrorResponse(err))
			return
		}

		url, err := svc.ShortenURL(r.Context(), req.URL)
		if err != nil {
			logger.Error(
				"failed to shorten url",
				slog.Group(op, slog.String("url", req.URL), slog.Any("err", err)),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, response.SuccessResponse(successMsg, toURLResponse(url)))
	}
}

func handleResolveShortCode(logger *slog.Logger, svc URLService) http.HandlerFunc {
	const op = "api.http.handleResolveShortCode"
	const successMsg = "The short code was successfully resolved."

	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := chi.URLParam(r, "shortCode")

		url, err := svc.ResolveShortCode(r.Context(), shortCode)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to resolve short code",
				slog.Group(op, slog.String("short_code", shortCode), slog.Any("err", err)),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.SuccessResponse(successMsg, toURLResponse(url)))
	}
}

func handleModifyURL(logger *slog.Logger, svc URLService, validate *validator.Validate) http.HandlerFunc {
	const op = "api.http.handleModifyURL"
	const successMsg = "The URL was successfully modified."

	return func(w http.ResponseWriter, r *http.Request) {
		var req urlRequest

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			if errors.Is(err, io.EOF) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.EmptyRequestBodyResponse)
				return
			}

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.BadRequestResponse)
			return
		}

		if err := validate.Struct(req); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationErrorResponse(err))
			return
		}

		shortCode := chi.URLParam(r, "shortCode")

		url, err := svc.ModifyURL(r.Context(), shortCode, req.URL)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to modify url",
				slog.Group(op, slog.String("short_code", shortCode), slog.String("url", req.URL), slog.Any("err", err)),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.SuccessResponse(successMsg, toURLResponse(url)))
	}
}

func handleDeactivateURL(logger *slog.Logger, svc URLService) http.HandlerFunc {
	const op = "api.http.handleDeactivateURL"
	const successMsg = "The URL was successfully deactivated."

	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := chi.URLParam(r, "shortCode")

		err := svc.DeactivateURL(r.Context(), shortCode)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to deactivate url",
				slog.Group(op, slog.String("short_code", shortCode), slog.Any("err", err)),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.SuccessResponse(successMsg))
	}
}

func handleGetURLStats(logger *slog.Logger, svc URLService) http.HandlerFunc {
	const op = "api.http.handleGetURLStats"
	const successMsg = "The URL statistics retrieved successfully."

	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := chi.URLParam(r, "shortCode")

		url, err := svc.GetURLStats(r.Context(), shortCode)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to get url stats",
				slog.Group(op, slog.String("short_code", shortCode), slog.Any("err", err)),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		data := toURLResponse(url)
		data.AccessCount = url.AccessCount

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.SuccessResponse(successMsg, data))
	}
}
