package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/todos"
	todosweb "go-echo-server-template/internal/todos/web"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { _ = logging.Initialize("test", "ERROR") }

func TestTodosListAndCreate(t *testing.T) {
	store := todos.NewFakeStore()
	svc := todos.NewService(store, nil, true)
	e := echo.New()
	e.HTTPErrorHandler = httpx.ErrorHandler
	r := httpx.NewRenderer("../../../web/templates")
	httpx.SetDefaultRenderer(r)
	e.Renderer = r
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := authn.WithPrincipal(c.Request().Context(), authn.Principal{UserID: "anonymous", AuthMethod: "disabled"})
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	todosweb.Register(e, svc, config.AuthConfig{CSRFCookieName: "csrf_token", Disabled: true})

	req := httptest.NewRequest(http.MethodGet, "/app/todos", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Todos")
	assert.Contains(t, rec.Body.String(), "No todos yet")

	req = httptest.NewRequest(http.MethodPost, "/app/todos", strings.NewReader("label=Hello"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Hello")
	assert.NotContains(t, rec.Body.String(), "No todos yet")

	req = httptest.NewRequest(http.MethodPost, "/app/todos", strings.NewReader("label="))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
