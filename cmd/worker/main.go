package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-echo-server-template/internal/app"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	cfg.Jobs.EmbedInServer = true
	cfg.Jobs.Enabled = true

	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		_ = application.Shutdown(shutdownCtx)
	}()

	logging.WithContext(ctx, "worker").Info("Running worker process", nil)
	application.StartWorkerOnly(ctx)
}
