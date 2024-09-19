package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/pkg/postgres"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg *config.Config) error {
	const op = "app.Run"

	db, err := postgres.New(
		ctx,
		cfg.Postgres.DSN(),
		postgres.WithConnMaxIdleTime(cfg.Postgres.ConnMaxIdleTime),
		postgres.WithConnMaxLifetime(cfg.Postgres.ConnMaxLifetime),
		postgres.WithMaxIdleConns(cfg.Postgres.MaxIdleConns),
		postgres.WithMaxOpenConns(cfg.Postgres.MaxOpenConns),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}
	defer db.Close()

	if err := postgres.RunMigrations("file://migrations", cfg.Postgres.DSN()); err != nil {
		return fmt.Errorf("%s: failed to run migrations: %w", op, err)
	}

	server := &http.Server{
		Addr:           cfg.HTTPServer.Addr(),
		Handler:        nil,
		ReadTimeout:    cfg.HTTPServer.ReadTimeout,
		WriteTimeout:   cfg.HTTPServer.WriteTimeout,
		IdleTimeout:    cfg.HTTPServer.IdleTimeout,
		MaxHeaderBytes: cfg.HTTPServer.MaxHeaderBytes,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error

		switch cfg.Env {
		case config.EnvProd:
			err = server.ListenAndServeTLS(cfg.HTTPServer.CertFile, cfg.HTTPServer.KeyFile)
		default:
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("%s: server error occurred: %w", op, err)
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		if err := server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("%s: failed to shutdown server: %w", op, err)
		}

		return nil
	})

	return g.Wait()
}
