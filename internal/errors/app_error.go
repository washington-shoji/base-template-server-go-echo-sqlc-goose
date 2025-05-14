package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// AppError represents a custom application error
type AppError struct {
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	ErrorCode  string                 `json:"error_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
	err        error                  // Original error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.err
}

// Common error types
var (
	ErrNotFound     = fmt.Errorf("resource not found")
	ErrBadRequest   = fmt.Errorf("bad request")
	ErrValidation   = fmt.Errorf("validation error")
	ErrInternal     = fmt.Errorf("internal server error")
	ErrUnauthorized = fmt.Errorf("unauthorized")
)

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusNotFound,
		ErrorCode:  "NOT_FOUND",
		Details:    details,
		err:        ErrNotFound,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusBadRequest,
		ErrorCode:  "BAD_REQUEST",
		Details:    details,
		err:        ErrBadRequest,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
		ErrorCode:  "VALIDATION_ERROR",
		Details:    details,
		err:        ErrValidation,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  "INTERNAL_ERROR",
		err:        fmt.Errorf("%w: %v", ErrInternal, err),
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		ErrorCode:  "UNAUTHORIZED",
		err:        ErrUnauthorized,
	}
}

// HTTPErrorHandler returns a custom error handler for Echo
func HTTPErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var appErr *AppError
		if !c.Response().Committed {
			if err == echo.ErrNotFound {
				appErr = NewNotFoundError("Resource not found", nil)
			} else if ae, ok := err.(*AppError); ok {
				appErr = ae
			} else {
				appErr = NewInternalError("An unexpected error occurred", err)
			}

			if err := c.JSON(appErr.StatusCode, appErr); err != nil {
				c.Logger().Error(err)
			}
		}
	}
}

// IsAppError checks if an error is an AppError and returns it
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
