package http

import (
	"context"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

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

func NewRouter(logger *slog.Logger, urlSvc URLService) *http.ServeMux {
	validate := getValidate()
	mux := http.NewServeMux()

	mux.Handle("POST /shorten", handleShortenURL(logger, urlSvc, validate))
	mux.Handle("GET /shorten/{shortCode}", handleResolveShortCode(logger, urlSvc))
	mux.Handle("PUT /shorten/{shortCode}", handleModifyURL(logger, urlSvc, validate))
	mux.Handle("DELETE /shorten/{shortCode}", handleDeactivateURL(logger, urlSvc))
	mux.Handle("GET /shorten/{shortCode}/stats", handleGetURLStats(logger, urlSvc))

	return mux
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
