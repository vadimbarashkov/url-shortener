package http

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/pkg/render"
)

var validate = validator.New()

type request struct {
	URL string `json:"url" validate:"required,url"`
}

type response struct {
	ID        int64     `json:"id"`
	ShortCode string    `json:"short_code"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

		resp := response{
			ID:        url.ID,
			ShortCode: url.ShortCode,
			URL:       url.OriginalURL,
			CreatedAt: url.CreatedAt,
			UpdatedAt: url.UpdatedAt,
		}

		if err := render.JSON(w, http.StatusCreated, resp); err != nil {
			logger.Error(
				"failed to render JSON response",
				slog.Group(op, slog.Any("err", err)),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}
