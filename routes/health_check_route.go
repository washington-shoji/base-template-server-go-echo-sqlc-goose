package routes

import (
	"context"
	"go-echo-server-template/handlers"
	"go-echo-server-template/internal/database"

	"github.com/labstack/echo/v4"
)

// HealthCheckRoutes sets up the health check routes
func HealthCheckRoutes(e *echo.Echo, ctx context.Context, db *database.Queries) {

	healthCheckHandler := handlers.NewHealthCheckHandler()

	e.GET("/health", healthCheckHandler.CheckHealth)
	e.GET("/", healthCheckHandler.CheckHealth) // Also handle root path as health check
}
