package http

import (
	"context"
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

type URLService interface {
	ShortenURL(ctx context.Context, originalURL string) (*models.URL, error)
	ResolveShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	ModifyURL(ctx context.Context, shortCode, originalURL string) (*models.URL, error)
	DeactivateURL(ctx context.Context, shortCode string) error
	GetURLStats(ctx context.Context, shortCode string) (*models.URL, error)
}

func getValidate() *validator.Validate {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return validate
}

func NewRouter(logger *slog.Logger, urlSvc URLService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept"},
		AllowCredentials: false,
		MaxAge:           84600,
	}))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		validate := getValidate()

		r.Get("/ping", handlePing)

		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", handleShortenURL(logger, urlSvc, validate))

			r.Route("/{shortCode}", func(r chi.Router) {
				r.Get("/", handleResolveShortCode(logger, urlSvc))
				r.Put("/", handleModifyURL(logger, urlSvc, validate))
				r.Delete("/", handleDeactivateURL(logger, urlSvc))
				r.Get("/stats", handleGetURLStats(logger, urlSvc))
			})
		})
	})

	return r
}
