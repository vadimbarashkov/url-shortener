package http

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/go-playground/validator/v10"
)

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
