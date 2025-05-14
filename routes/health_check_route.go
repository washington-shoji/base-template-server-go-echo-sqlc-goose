package routes

import (
	"context"
	"go-echo-server-template/handlers"
	"go-echo-server-template/internal/database"

	"github.com/labstack/echo/v4"
)

func HealthCheckRoutes(e *echo.Echo, ctx context.Context, db *database.Queries) {

	healthCheckHandler := handlers.NewFavoriteCoinsHandler()

	e.GET("/", healthCheckHandler.ServerHealthCheck)
}
