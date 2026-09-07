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

	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := application.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	logging.WithContext(shutdownCtx, "shutdown").Info("Shutting down", nil)
	if err := application.Shutdown(shutdownCtx); err != nil {
		logging.WithContext(shutdownCtx, "shutdown").Error("shutdown error", err, nil)
	}
}
