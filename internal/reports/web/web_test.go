package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/reports"
	reportsweb "go-echo-server-template/internal/reports/web"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { _ = logging.Initialize("test", "ERROR") }

func setupEcho(t *testing.T, svc *reports.Service) *echo.Echo {
	t.Helper()
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
	reportsweb.Register(e, svc, config.AuthConfig{CSRFCookieName: "csrf_token", Disabled: true})
	return e
}

func TestListAndStatusFragments(t *testing.T) {
	store := reports.NewFakeStore()
	svc := reports.NewService(store, reports.FakeSummary{Value: reports.TodoSummary{Total: 1}}, &reports.FakeEnqueuer{}, true, time.Minute, time.Minute)
	e := setupEcho(t, svc)

	created, err := svc.CreateReport(
		authn.WithPrincipal(httptest.NewRequest(http.MethodGet, "/", nil).Context(), authn.Principal{UserID: "anonymous"}),
		reports.CreateRequest{Type: reports.TypeTodoSummary},
	)
	require.NoError(t, err)
	require.NoError(t, svc.Generate(
		authn.WithPrincipal(httptest.NewRequest(http.MethodGet, "/", nil).Context(), authn.Principal{UserID: "system"}),
		created.ReportID,
	))

	req := httptest.NewRequest(http.MethodGet, "/app/reports", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Reports")

	req = httptest.NewRequest(http.MethodGet, "/app/reports", nil)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "todo-summary")

	req = httptest.NewRequest(http.MethodGet, "/app/reports/"+created.ReportID.String()+"/status", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Completed")
	assert.NotContains(t, rec.Body.String(), `hx-trigger="every 2s"`)

	req = httptest.NewRequest(http.MethodPost, "/app/reports", strings.NewReader("type=todo-summary"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStatusPollTriggerOnlyNonTerminal(t *testing.T) {
	store := reports.NewFakeStore()
	svc := reports.NewService(store, reports.FakeSummary{}, &reports.FakeEnqueuer{}, true, time.Minute, time.Minute)
	e := setupEcho(t, svc)
	now := time.Now().UTC()

	cases := []struct {
		status string
		poll   bool
	}{
		{reports.StatusPending, true},
		{reports.StatusProcessing, true},
		{reports.StatusCompleted, false},
		{reports.StatusFailed, false},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			id := uuid.New()
			store.Seed(reports.Report{
				ReportID: id, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
				Status: tc.status, CreatedAt: now, UpdatedAt: now,
			})
			req := httptest.NewRequest(http.MethodGet, "/app/reports/"+id.String()+"/status", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			body := rec.Body.String()
			if tc.poll {
				assert.Contains(t, body, `hx-trigger="every 2s"`)
			} else {
				assert.NotContains(t, body, `hx-trigger`)
			}
		})
	}
}
