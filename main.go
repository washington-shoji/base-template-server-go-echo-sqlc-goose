package main

import (
	"context"
	"fmt"
	"go-echo-server-template/internal/logger"
	"go-echo-server-template/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	e, db, err := server.Start()
	if err != nil {
		// Use a basic fmt.Printf here as logger might not be initialized yet
		// or if server.Start() failed before logger initialization.
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit // Block until a signal is received

	logger.WithContext(context.Background(), "shutdown").Info("Shutting down server...", nil)

	// Create a context with a timeout for the shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the Echo server.
	if err := e.Shutdown(ctx); err != nil {
		logger.WithContext(context.Background(), "shutdown").Error("Server shutdown failed", err, nil)
	} else {
		logger.WithContext(context.Background(), "shutdown").Info("Server gracefully stopped.", nil)
	}

	// Close the database connection.
	if db != nil {
		if err := db.Close(); err != nil {
			logger.WithContext(context.Background(), "shutdown").Error("Failed to close database connection", err, nil)
		} else {
			logger.WithContext(context.Background(), "shutdown").Info("Database connection closed.", nil)
		}
	}
}
