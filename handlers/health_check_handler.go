package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthCheckHandler handles health check requests
type HealthCheckHandler struct{}

// NewHealthCheckHandler creates a new HealthCheckHandler
func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

// CheckHealth handles the /health endpoint
func (h *HealthCheckHandler) CheckHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Server is healthy",
	})
}
