// Package http provides the HTTP delivery layer for the URL shortener service.
// This package contains the HTTP handlers and related types used for processing
// incoming requests, validating input, and formatting responses.
package http

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/go-playground/validator/v10"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter initializes and returns a new Chi router configured with middleware and routes for the URL shortener API.
func NewRouter(logger *httplog.Logger, urlUseCase urlUseCase) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept"},
		AllowCredentials: false,
		MaxAge:           84600,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.yml"),
	))

	r.Get("/docs/swagger.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.yml")
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", handlePing)

		r.Route("/shorten", func(r chi.Router) {
			validate := validator.New()
			h := newURLHandler(urlUseCase, validate)

			r.Post("/", h.shortenURL)

			r.Route("/{shortCode}", func(r chi.Router) {
				r.Get("/", h.resolveShortCode)
				r.Put("/", h.modifyURL)
				r.Delete("/", h.deactivateURL)
				r.Get("/stats", h.getURLStats)
			})
		})
	})

	return r
}
