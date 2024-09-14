package http

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

// URLService defines the interface for the core URL shortening business logic.
type URLService interface {
	// ShortenURL creates a shortened version of the provided original URL.
	// It returns the generated short code and associated URL details, or an error if the operation fails.
	ShortenURL(ctx context.Context, originalURL string) (*models.URL, error)

	// ResolveShortCode retrieves the original URL for a given short code.
	// It returns the associated URL details or an error if the URL is not found.
	ResolveShortCode(ctx context.Context, shortCode string) (*models.URL, error)

	// ModifyURL updates the original URL linked to the provided short code.
	// It returns the modified URL details or an error if the operation fails or the URL doesn't exist.
	ModifyURL(ctx context.Context, shortCode, originalURL string) (*models.URL, error)

	// DeactivateURL disables the URL, making it no longer functional.
	// It returns an error if the URL doesn't exist or if deactivation fails.
	DeactivateURL(ctx context.Context, shortCode string) error

	// GetURLStats retrieves the statistics (e.g., access count) of the URL associated with the short code.
	// It returns the statistics or an error if the URL is not found.
	GetURLStats(ctx context.Context, shortCode string) (*models.URL, error)
}

// getValidate initializes a new validator instance for validating incoming request payloads.
// It customizes tag name extraction from struct fields to match JSON tags.
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

// NewRouter initializes and returns a new HTTP router with all routes and middleware configured.
func NewRouter(logger *httplog.Logger, urlSvc URLService) http.Handler {
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
		validate := getValidate()

		r.Get("/ping", handlePing)

		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", handleShortenURL(urlSvc, validate))

			r.Route("/{shortCode}", func(r chi.Router) {
				r.Get("/", handleResolveShortCode(urlSvc))
				r.Put("/", handleModifyURL(urlSvc, validate))
				r.Delete("/", handleDeactivateURL(urlSvc))
				r.Get("/stats", handleGetURLStats(urlSvc))
			})
		})
	})

	return r
}
