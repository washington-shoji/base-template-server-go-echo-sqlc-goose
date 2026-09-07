package httpx

import (
	"errors"
	"html"
	"net/http"
	"strings"

	"go-echo-server-template/internal/platform/authz"
	"go-echo-server-template/internal/platform/logging"

	"github.com/labstack/echo/v4"
)

// Domain error sentinels used across modules. Adapters map these to HTTP/HTML.
var (
	ErrNotFound         = errors.New("not found")
	ErrValidation       = errors.New("validation error")
	ErrBadRequest       = errors.New("bad request")
	ErrConflict         = errors.New("conflict")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrPermissionDenied = authz.ErrPermissionDenied
	ErrInternal         = errors.New("internal error")
)

type ErrorBody struct {
	Status  int                    `json:"status"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// APIError is an adapter-level error with optional details (not used in domain packages).
type APIError struct {
	Err     error
	Message string
	Details map[string]interface{}
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "error"
}

func (e *APIError) Unwrap() error { return e.Err }

func NewAPIError(err error, message string, details map[string]interface{}) *APIError {
	return &APIError{Err: err, Message: message, Details: details}
}

func MapError(err error) (int, string, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "Resource not found"
	case errors.Is(err, ErrValidation):
		return http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed"
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, "BAD_REQUEST", "Bad request"
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, "CONFLICT", "Conflict"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized"
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden, "FORBIDDEN", "Permission denied"
	default:
		return http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred"
	}
}

func errorPage(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "pages/errors/400.html"
	case http.StatusUnauthorized:
		return "pages/errors/401.html"
	case http.StatusForbidden:
		return "pages/errors/403.html"
	case http.StatusNotFound:
		return "pages/errors/404.html"
	case http.StatusUnprocessableEntity:
		return "pages/errors/422.html"
	default:
		if status >= 500 {
			return "pages/errors/500.html"
		}
		return "pages/errors/500.html"
	}
}

func ErrorHandler(err error, c echo.Context) {
	log := logging.WithContext(c.Request().Context(), "error_handler")

	var (
		status  int
		code    string
		message string
		details map[string]interface{}
	)

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		status, code, message = MapError(apiErr.Err)
		if apiErr.Message != "" {
			message = apiErr.Message
		}
		details = apiErr.Details
	} else if he, ok := err.(*echo.HTTPError); ok {
		status = he.Code
		code = "HTTP_ERROR"
		switch m := he.Message.(type) {
		case string:
			message = m
		default:
			message = http.StatusText(he.Code)
		}
	} else {
		status, code, message = MapError(err)
		if status != http.StatusInternalServerError && err != nil {
			message = err.Error()
		}
	}

	log.Error("Request error", err, map[string]interface{}{
		"error_code":  code,
		"method":      c.Request().Method,
		"path":        c.Request().URL.Path,
		"status_code": status,
	})

	if c.Response().Committed {
		return
	}

	safe := html.EscapeString(message)
	accept := c.Request().Header.Get("Accept")
	wantsHTML := IsHTMX(c) || (accept != "" && accept != "application/json" && !isAPI(c.Path()))
	if wantsHTML {
		if IsHTMX(c) {
			_ = c.HTML(status, `<div class="rounded-md border border-danger/40 bg-danger/10 px-3 py-2 text-sm text-danger" role="alert">`+safe+`</div>`)
			return
		}
		data := struct {
			Shell
			Message string
		}{
			Shell:   Shell{Title: http.StatusText(status)},
			Message: message, // template escapes
		}
		_ = RenderAppStatus(c, status, errorPage(status), data)
		return
	}

	_ = c.JSON(status, ErrorBody{Status: status, Code: code, Message: message, Details: details})
}

func isAPI(path string) bool {
	return strings.HasPrefix(path, "/api")
}
