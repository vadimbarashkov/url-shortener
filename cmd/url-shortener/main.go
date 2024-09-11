package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/httplog/v2"
	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/internal/database/postgres"
	"github.com/vadimbarashkov/url-shortener/internal/service"
	"golang.org/x/sync/errgroup"

	myhttp "github.com/vadimbarashkov/url-shortener/internal/api/http"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}

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

	r := myhttp.NewRouter(httplog.NewLogger(""), urlSvc)

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
		err := server.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile)
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
