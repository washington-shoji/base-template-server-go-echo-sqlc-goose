package errors

import (
	"go-echo-server-template/internal/logger"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Status  int                    `json:"status"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorHandler is a custom error handler for Echo
func ErrorHandler(err error, c echo.Context) {
	log := logger.WithContext(c.Request().Context(), "error_handler")

	var response *ErrorResponse
	var httpStatus int

	// Check if it's our custom AppError
	if appErr, ok := IsAppError(err); ok {
		response = &ErrorResponse{
			Status:  appErr.StatusCode,
			Code:    appErr.ErrorCode,
			Message: appErr.Message,
			Details: appErr.Details,
		}
		httpStatus = appErr.StatusCode
	} else {
		// Handle Echo's HTTPError
		if he, ok := err.(*echo.HTTPError); ok {
			response = &ErrorResponse{
				Status:  he.Code,
				Code:    "HTTP_ERROR",
				Message: he.Message.(string),
			}
			httpStatus = he.Code
		} else {
			// Handle unknown errors
			response = &ErrorResponse{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "An unexpected error occurred",
			}
			httpStatus = http.StatusInternalServerError
		}
	}

	// Log the error with context
	log.Error("Request error", err, map[string]interface{}{
		"status_code": response.Status,
		"error_code":  response.Code,
		"path":        c.Request().URL.Path,
		"method":      c.Request().Method,
	})

	// Don't send error details in production
	if c.Echo().Debug {
		response.Details = map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Send response
	if !c.Response().Committed {
		if err := c.JSON(httpStatus, response); err != nil {
			log.Error("Failed to send error response", err, nil)
		}
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationErrors creates a validation error response
func NewValidationErrors(errors []ValidationError) *AppError {
	details := make(map[string]interface{})
	for _, err := range errors {
		details[err.Field] = err.Message
	}
	return NewValidationError("Validation failed", details)
}
