package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/reports"
	reportsapi "go-echo-server-template/internal/reports/api"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { _ = logging.Initialize("test", "ERROR") }

func newSvc(store reports.Store, sum reports.TodoSummaryReader, enq reports.Enqueuer) *reports.Service {
	return reports.NewService(store, sum, enq, false, time.Minute, time.Minute)
}

func withUser(e *echo.Echo, uid string, svc *reports.Service) {
	e.HTTPErrorHandler = httpx.ErrorHandler
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := authn.WithPrincipal(c.Request().Context(), authn.Principal{UserID: uid, AuthMethod: "test"})
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	reportsapi.Register(e, svc)
}

func TestCreateAcceptedAndGet(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{Value: reports.TodoSummary{Total: 2}}, &reports.FakeEnqueuer{})
	e := echo.New()
	withUser(e, "u1", svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports", strings.NewReader(`{"type":"todo-summary"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)
	assert.Contains(t, rec.Header().Get(echo.HeaderLocation), "/api/v1/reports/")

	var created reports.Report
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.Equal(t, reports.StatusPending, created.Status)
	assert.Equal(t, rec.Header().Get(echo.HeaderLocation), "/api/v1/reports/"+created.ReportID.String())

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reports/"+created.ReportID.String(), nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOwnershipForbidden(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{}, &reports.FakeEnqueuer{})

	owner := echo.New()
	withUser(owner, "owner", svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports", strings.NewReader(`{"type":"todo-summary"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	owner.ServeHTTP(rec, req)
	var created reports.Report
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	intruder := echo.New()
	withUser(intruder, "intruder", svc)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/reports/"+created.ReportID.String(), nil)
	rec = httptest.NewRecorder()
	intruder.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
