package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
	"github.com/vadimbarashkov/url-shortener/pkg/response"
)

// handlePing handles health check requests to ensure the server is running.
func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "pong")
}

// urlRequest represents the request payload for creating or updating a shortened URL.
type urlRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// urlResponse represents the response payload for a shortened URL operation.
type urlResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	URL         string    `json:"url"`
	AccessCount int64     `json:"access_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toURLResponse converts a URL model from the business layer into a response payload.
func toURLResponse(url *models.URL) urlResponse {
	return urlResponse{
		ID:        url.ID,
		ShortCode: url.ShortCode,
		URL:       url.OriginalURL,
		CreatedAt: url.CreatedAt,
		UpdatedAt: url.UpdatedAt,
	}
}

// handleShortenURL handles POST requests to shorten a URL.
//
// The request must contain a valid URL. The handler validates the input, calls the URL shortening
// service, and returns the generated short code with relevant metadata.
func handleShortenURL(svc URLService, validate *validator.Validate) http.HandlerFunc {
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
			httplog.LogEntrySetFields(r.Context(), map[string]any{"op": op, "err": err})

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, response.SuccessResponse(successMsg, toURLResponse(url)))
	}
}

// handleResolveShortCode handles GET requests to resolve a short code into the original URL.
//
// The handler fetches the original URL based on the provided short code, returning
// the URL data if found or a 404 error if not.
func handleResolveShortCode(svc URLService) http.HandlerFunc {
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

			httplog.LogEntrySetFields(r.Context(), map[string]any{"op": op, "err": err})

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.SuccessResponse(successMsg, toURLResponse(url)))
	}
}

// handleModifyURL handles PUT requests to modify an existing URL.
//
// The request must contain a valid new URL. The handler updates the URL with the new URL,
// returning the updated URL metadata.
func handleModifyURL(svc URLService, validate *validator.Validate) http.HandlerFunc {
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

			httplog.LogEntrySetFields(r.Context(), map[string]any{"op": op, "err": err})

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.SuccessResponse(successMsg, toURLResponse(url)))
	}
}

// handleDeactivateURL handles DELETE requests to deactivate the URL.
//
// Once deactivated, the URL will no longer be functional. The handler returns a success message
// if deactivation is successful or an error if the short code doesn't exist.
func handleDeactivateURL(svc URLService) http.HandlerFunc {
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

			httplog.LogEntrySetFields(r.Context(), map[string]any{"op": op, "err": err})

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ServerErrorResponse)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.SuccessResponse(successMsg))
	}
}

// handleGetURLStats handles GET requests to retrieve usage statistics for a shortened URL.
//
// The handler fetches access counts and other statistics for the given shortened URL, returning the data
// or a 404 error if the URL doesn't exist.
func handleGetURLStats(svc URLService) http.HandlerFunc {
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

			httplog.LogEntrySetFields(r.Context(), map[string]any{"op": op, "err": err})

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
