package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/vadimbarashkov/url-shortener/internal/app"
	"github.com/vadimbarashkov/url-shortener/internal/config"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
