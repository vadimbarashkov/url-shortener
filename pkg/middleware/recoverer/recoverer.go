package recoverer

import (
	"log/slog"
	"net/http"

	"github.com/vadimbarashkov/url-shortener/pkg/middleware"
	"github.com/vadimbarashkov/url-shortener/pkg/render"
	"github.com/vadimbarashkov/url-shortener/pkg/response"
)

func New(logger *slog.Logger) middleware.Middleware {
	const op = "middleware.recoverer.New"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(
						"something went wrong, panic occuried",
						slog.Group(op, slog.Any("err", err)),
					)

					render.JSON(w, http.StatusInternalServerError, response.ServerErrorResponse)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
