package http

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
	"github.com/vadimbarashkov/url-shortener/pkg/render"

	stdresp "github.com/vadimbarashkov/url-shortener/pkg/response"
)

var validate = validator.New()

type request struct {
	URL string `json:"url" validate:"required,url"`
}

type response struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	URL         string    `json:"url"`
	AccessCount *int64    `json:"access_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toResponse(url *models.URL) response {
	return response{
		ID:        url.ID,
		ShortCode: url.ShortCode,
		URL:       url.OriginalURL,
		CreatedAt: url.CreatedAt,
		UpdatedAt: url.UpdatedAt,
	}
}

func handleShortenURL(logger *slog.Logger, svc URLService) http.Handler {
	const op = "api.http.handleShortenURL"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req request

		if err := render.BindJSON(r, &req); err != nil {
			if errors.Is(err, io.EOF) {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			logger.Error(
				"failed to parse JSON from request body",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := validate.Struct(req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		url, err := svc.ShortenURL(r.Context(), req.URL)
		if err != nil {
			logger.Error(
				"failed to shorten url",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := toResponse(url)
		resp := stdresp.SuccessResponse(http.StatusCreated, "The URL has been shortened successfully.", data)

		render.JSON(w, http.StatusCreated, resp)
	})
}

func handleResolveShortCode(logger *slog.Logger, svc URLService) http.Handler {
	const op = "api.http.handleResolveShortCode"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("shortCode")

		url, err := svc.ResolveShortCode(r.Context(), shortCode)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			logger.Error(
				"failed to resolve short code",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := toResponse(url)
		resp := stdresp.SuccessResponse(http.StatusOK, "The short code was successfully resolved.", data)

		render.JSON(w, http.StatusOK, resp)
	})
}

func handleModifyURL(logger *slog.Logger, svc URLService) http.Handler {
	const op = "api.http.handleModifyURL"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("shortCode")

		var req request

		if err := render.BindJSON(r, &req); err != nil {
			if errors.Is(err, io.EOF) {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			logger.Error(
				"failed to parse JSON from request body",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := validate.Struct(req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		url, err := svc.ModifyURL(r.Context(), shortCode, req.URL)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			logger.Error(
				"failed to modify url",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := toResponse(url)
		resp := stdresp.SuccessResponse(http.StatusOK, "The URL was successfully modified.", data)

		render.JSON(w, http.StatusOK, resp)
	})
}

func handleDeactivateURL(logger *slog.Logger, svc URLService) http.Handler {
	const op = "api.http.handleDeactivateURL"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("shortCode")

		err := svc.DeactivateURL(r.Context(), shortCode)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			logger.Error(
				"failed to deactivate url",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		resp := stdresp.SuccessResponse(http.StatusOK, "The URL was successfully deactivated.")

		render.JSON(w, http.StatusOK, resp)
	})
}

func handleGetURLStats(logger *slog.Logger, svc URLService) http.Handler {
	const op = "api.http.handleGetURLStats"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("shortCode")

		url, err := svc.GetURLStats(r.Context(), shortCode)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			logger.Error(
				"failed to get url stats",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := toResponse(url)
		data.AccessCount = &url.AccessCount
		resp := stdresp.SuccessResponse(http.StatusOK, "URL statistics retrieved successfully.", data)

		render.JSON(w, http.StatusOK, resp)
	})
}
