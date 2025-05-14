package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common error types that can be used across the application
var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrForbidden         = errors.New("forbidden access")
	ErrInternalServer    = errors.New("internal server error")
	ErrConflict          = errors.New("resource conflict")
	ErrBadRequest        = errors.New("bad request")
	ErrValidation        = errors.New("validation error")
	ErrDatabaseOperation = errors.New("database operation error")
)

// AppError represents a custom error type that includes additional context
type AppError struct {
	Err        error
	Message    string
	StatusCode int
	Code       string
	Details    map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap implements the unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(err error, message string, statusCode int, code string, details map[string]interface{}) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
		Code:       code,
		Details:    details,
	}
}

// Error constructors for common scenarios
func NewNotFoundError(message string, details map[string]interface{}) *AppError {
	return NewAppError(ErrNotFound, message, http.StatusNotFound, "NOT_FOUND", details)
}

func NewBadRequestError(message string, details map[string]interface{}) *AppError {
	return NewAppError(ErrBadRequest, message, http.StatusBadRequest, "BAD_REQUEST", details)
}

func NewValidationError(message string, details map[string]interface{}) *AppError {
	return NewAppError(ErrValidation, message, http.StatusUnprocessableEntity, "VALIDATION_ERROR", details)
}

func NewInternalServerError(err error, message string) *AppError {
	return NewAppError(err, message, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", nil)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrUnauthorized, message, http.StatusUnauthorized, "UNAUTHORIZED", nil)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(ErrForbidden, message, http.StatusForbidden, "FORBIDDEN", nil)
}

func NewConflictError(message string, details map[string]interface{}) *AppError {
	return NewAppError(ErrConflict, message, http.StatusConflict, "CONFLICT", details)
}

// IsAppError checks if an error is an AppError and returns it
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
