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

func handleShortenURL(logger *slog.Logger, svc URLService, validate *validator.Validate) http.Handler {
	const op = "api.http.handleShortenURL"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req urlRequest

		if err := render.BindJSON(r, &req); err != nil {
			if errors.Is(err, io.EOF) {
				render.JSON(w, http.StatusBadRequest, stdresp.EmptyRequestBodyResponse)
				return
			}

			logger.Error(
				"failed to parse JSON from request body",
				slog.Group(op, slog.Any("err", err)),
			)

			render.JSON(w, http.StatusInternalServerError, stdresp.ServerErrorResponse)
			return
		}

		if err := validate.Struct(req); err != nil {
			render.JSON(w, http.StatusBadRequest, stdresp.ValidationErrorResponse(err))
			return
		}

		url, err := svc.ShortenURL(r.Context(), req.URL)
		if err != nil {
			logger.Error(
				"failed to shorten url",
				slog.Group(op, slog.String("url", req.URL), slog.Any("err", err)),
			)

			render.JSON(w, http.StatusInternalServerError, stdresp.ServerErrorResponse)
			return
		}

		data := toURLResponse(url)
		resp := stdresp.SuccessResponse("The URL has been shortened successfully.", data)

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
				render.JSON(w, http.StatusNotFound, stdresp.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to resolve short code",
				slog.Group(op, slog.String("short_code", shortCode), slog.Any("err", err)),
			)

			render.JSON(w, http.StatusInternalServerError, stdresp.ServerErrorResponse)
			return
		}

		data := toURLResponse(url)
		resp := stdresp.SuccessResponse("The short code was successfully resolved.", data)

		render.JSON(w, http.StatusOK, resp)
	})
}

func handleModifyURL(logger *slog.Logger, svc URLService, validate *validator.Validate) http.Handler {
	const op = "api.http.handleModifyURL"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("shortCode")

		var req urlRequest

		if err := render.BindJSON(r, &req); err != nil {
			if errors.Is(err, io.EOF) {
				render.JSON(w, http.StatusBadRequest, stdresp.EmptyRequestBodyResponse)
				return
			}

			logger.Error(
				"failed to parse JSON from request body",
				slog.Group(op, slog.Any("err", err)),
			)

			render.JSON(w, http.StatusInternalServerError, stdresp.ServerErrorResponse)
			return
		}

		if err := validate.Struct(req); err != nil {
			render.JSON(w, http.StatusBadRequest, stdresp.ValidationErrorResponse(err))
			return
		}

		url, err := svc.ModifyURL(r.Context(), shortCode, req.URL)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				render.JSON(w, http.StatusNotFound, stdresp.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to modify url",
				slog.Group(op, slog.String("short_code", shortCode), slog.String("url", req.URL), slog.Any("err", err)),
			)

			render.JSON(w, http.StatusInternalServerError, stdresp.ServerErrorResponse)
			return
		}

		data := toURLResponse(url)
		resp := stdresp.SuccessResponse("The URL was successfully modified.", data)

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
				render.JSON(w, http.StatusNotFound, stdresp.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to deactivate url",
				slog.Group(op, slog.String("short_code", shortCode), slog.Any("err", err)),
			)

			render.JSON(w, http.StatusInternalServerError, stdresp.ServerErrorResponse)
			return
		}

		resp := stdresp.SuccessResponse("The URL was successfully deactivated.")

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
				render.JSON(w, http.StatusNotFound, stdresp.ResourceNotFoundResponse)
				return
			}

			logger.Error(
				"failed to get jurl stats",
				slog.Group(op, slog.String("short_code", shortCode), slog.Any("err", err)),
			)

			render.JSON(w, http.StatusInternalServerError, stdresp.ServerErrorResponse)
			return
		}

		data := toURLResponse(url)
		data.AccessCount = url.AccessCount
		resp := stdresp.SuccessResponse("URL statistics retrieved successfully.", data)

		render.JSON(w, http.StatusOK, resp)
	})
}
