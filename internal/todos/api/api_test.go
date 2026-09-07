package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/todos"
	todoapi "go-echo-server-template/internal/todos/api"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = logging.Initialize("test", "ERROR")
}

func setup() (*echo.Echo, *todos.Service) {
	e := echo.New()
	e.HTTPErrorHandler = httpx.ErrorHandler
	svc := todos.NewService(todos.NewFakeStore(), nil, true)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := authn.WithPrincipal(c.Request().Context(), authn.Principal{UserID: "u", AuthMethod: "disabled"})
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	todoapi.Register(e, svc)
	return e, svc
}

func TestCreateAndLegacyRoutes(t *testing.T) {
	e, _ := setup()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(`{"label":"a"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var todo todos.Todo
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &todo))

	req = httptest.NewRequest(http.MethodGet, "/api/v1/todo/"+todo.TodoID.String(), nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/create-todo", strings.NewReader(`{"label":""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
