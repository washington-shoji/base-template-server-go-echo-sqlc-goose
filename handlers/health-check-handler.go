package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type healthCheckResponse struct {
	Response string `json:"response"`
}

type HealthCheckHandler struct {
}

func NewFavoriteCoinsHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

func NewServerHealthCheckHandler() {

}

func (h *HealthCheckHandler) ServerHealthCheck(c echo.Context) error {
	result := healthCheckResponse{Response: "Server is alive and kicking!!!"}
	return c.JSON(http.StatusOK, result)
}
