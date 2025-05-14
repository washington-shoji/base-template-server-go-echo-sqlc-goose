package logger

import (
	"github.com/labstack/echo/v4"
)

// LoggerMiddleware creates a new logger middleware
func LoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a new logger instance for this request
			log := WithContext(c.Request().Context(), "http")

			// Add logger to context
			c.Set("logger", log)

			// Log request
			log.Info("Incoming request", map[string]interface{}{
				"method":     c.Request().Method,
				"path":       c.Request().URL.Path,
				"remote_ip":  c.Request().RemoteAddr,
				"user_agent": c.Request().UserAgent(),
			})

			// Process request
			err := next(c)

			// Log response
			if err != nil {
				log.Error("Request failed", err, map[string]interface{}{
					"status_code": c.Response().Status,
				})
			} else {
				log.Info("Request completed", map[string]interface{}{
					"status_code": c.Response().Status,
				})
			}

			return err
		}
	}
}
