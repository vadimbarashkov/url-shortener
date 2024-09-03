package http

import (
	"context"
	"log/slog"
	"net/http"

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
	mux := http.NewServeMux()

	mux.Handle("POST /shorten", handleShortenURL(logger, urlSvc))
	mux.Handle("GET /shorten/{shortCode}", handleResolveShortCode(logger, urlSvc))
	mux.Handle("PUT /shorten/{shortCode}", handleModifyURL(logger, urlSvc))
	mux.Handle("DELETE /shorten/{shortCode}", handleDeactivateURL(logger, urlSvc))

	return mux
}
