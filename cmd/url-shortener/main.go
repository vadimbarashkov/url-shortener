package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/httplog/v2"
	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/internal/database/postgres"
	"github.com/vadimbarashkov/url-shortener/internal/service"
	"golang.org/x/sync/errgroup"

	api "github.com/vadimbarashkov/url-shortener/internal/api/http/v1"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const (
	// envDev is a constant string representing the development environment.
	envDev = "dev"
	// envStage is a constant string representing the staging environment.
	envStage = "stg"
	// envProd is a constant string representing the production environment.
	envProd = "prod"
)

// setupLogger configures and returns an HTTP logger based on the provided environment.
func setupLogger(env string) *httplog.Logger {
	opts := httplog.Options{
		LogLevel:        slog.LevelDebug,
		Concise:         true,
		RequestHeaders:  true,
		ResponseHeaders: true,
	}

	switch env {
	case envStage:
		opts.JSON = true
	case envProd:
		opts.LogLevel = slog.LevelInfo
		opts.JSON = true
	default:
		env = envDev
	}

	logger := httplog.NewLogger("url-shortener", opts)
	logger.Logger = logger.With(slog.String("env", env))

	return logger
}

// run initializes the application, sets up services, and starts the HTTP server.
func run(ctx context.Context) error {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		return err
	}
	g.Go(func() error {
		<-ctx.Done()
		return db.Close()
	})

	urlRepo := postgres.NewURLRepository(db)
	urlSvc := service.NewURLService(urlRepo, cfg.ShortCodeLength)

	logger := setupLogger(cfg.Env)
	r := api.NewRouter(logger, urlSvc)

	server := &http.Server{
		Addr:           cfg.Server.Addr(),
		Handler:        r,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g.Go(func() error {
		var err error

		switch cfg.Env {
		case envProd:
			err = server.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile)
		default:
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	return g.Wait()
}
