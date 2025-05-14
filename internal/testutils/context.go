package testutils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MockContext creates a new Echo context for testing
func MockContext(method, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// MockContextWithParams creates a new Echo context with path parameters for testing
func MockContextWithParams(method, path string, body string, params map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	ctx, rec := MockContext(method, path, body)

	// Add path parameters
	paramNames := make([]string, 0, len(params))
	paramValues := make([]string, 0, len(params))
	for name, value := range params {
		paramNames = append(paramNames, name)
		paramValues = append(paramValues, value)
	}

	ctx.SetParamNames(paramNames...)
	ctx.SetParamValues(paramValues...)

	return ctx, rec
}

// TestContext creates a new context with request ID for testing
func TestContext() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "request_id", uuid.New().String())
}

// MockRequest creates a new http.Request for testing
func MockRequest(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	return req.WithContext(TestContext())
}
