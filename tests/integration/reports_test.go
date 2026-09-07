package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go-echo-server-template/internal/app"
	"go-echo-server-template/internal/reports"
	reportspg "go-echo-server-template/internal/reports/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportsAsyncGeneration(t *testing.T) {
	body, _ := json.Marshal(reports.CreateRequest{Type: reports.TypeTodoSummary})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testEcho.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/api/v1/reports/")

	var created reports.Report
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.Equal(t, reports.StatusPending, created.Status)

	var got reports.Report
	require.Eventually(t, func() bool {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/"+created.ReportID.String(), nil)
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			return false
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		return got.Status == reports.StatusCompleted
	}, 5*time.Second, 200*time.Millisecond)

	assert.Contains(t, got.Result, "Total todos:")

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	rec = httptest.NewRecorder()
	testEcho.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestReportsOwnershipWithAuth(t *testing.T) {
	cfg := *testCfg
	cfg.Auth.Disabled = false
	cfg.Jobs.EmbedInServer = false
	cfg.Jobs.Enabled = true

	authApp, err := app.New(&cfg)
	require.NoError(t, err)
	defer authApp.DB.Close()

	register := func(email, password string) {
		body, _ := json.Marshal(map[string]string{"email": email, "password": password})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		authApp.Echo.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	}
	token := func(email, password string) string {
		body, _ := json.Marshal(map[string]string{"email": email, "password": password})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		authApp.Echo.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var out map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
		return out["token"]
	}

	register("a@example.com", "password123")
	register("b@example.com", "password123")
	tokA := token("a@example.com", "password123")
	tokB := token("b@example.com", "password123")

	body, _ := json.Marshal(reports.CreateRequest{Type: reports.TypeTodoSummary})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokA)
	rec := httptest.NewRecorder()
	authApp.Echo.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())
	var created reports.Report
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reports/"+created.ReportID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokB)
	rec = httptest.NewRecorder()
	authApp.Echo.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reports/"+created.ReportID.String(), nil)
	rec = httptest.NewRecorder()
	authApp.Echo.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reports/"+created.ReportID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokA)
	rec = httptest.NewRecorder()
	authApp.Echo.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestReportsConcurrentClaim(t *testing.T) {
	store := reportspg.NewStore(testDB)
	now := time.Now().UTC()
	id := uuid.New()
	_, err := store.CreateReport(context.Background(), reports.Report{
		ReportID: id, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
		Status: reports.StatusPending, CreatedAt: now, UpdatedAt: now,
	})
	require.NoError(t, err)

	svc := reports.NewService(
		store,
		reports.FakeSummary{Value: reports.TodoSummary{Total: 1}},
		&reports.FakeEnqueuer{},
		true,
		5*time.Minute,
		time.Minute,
	)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- svc.Generate(context.Background(), id)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	got, err := store.GetReportByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, reports.StatusCompleted, got.Status)
}

func TestReportsStaleProcessingReclaim(t *testing.T) {
	store := reportspg.NewStore(testDB)
	stale := time.Now().UTC().Add(-10 * time.Minute)
	id := uuid.New()
	_, err := store.CreateReport(context.Background(), reports.Report{
		ReportID: id, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
		Status: reports.StatusPending, CreatedAt: stale, UpdatedAt: stale,
	})
	require.NoError(t, err)
	_, err = store.ClaimExecution(context.Background(), id, stale, stale.Add(-time.Minute))
	require.NoError(t, err)

	svc := reports.NewService(
		store,
		reports.FakeSummary{Value: reports.TodoSummary{Total: 4}},
		&reports.FakeEnqueuer{},
		true,
		5*time.Minute,
		time.Minute,
	)
	require.NoError(t, svc.Generate(context.Background(), id))
	got, err := store.GetReportByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, reports.StatusCompleted, got.Status)
	assert.Contains(t, got.Result, "Total todos: 4")
}

func TestReportsReconcileReEnqueues(t *testing.T) {
	store := reportspg.NewStore(testDB)
	old := time.Now().UTC().Add(-2 * time.Minute)
	id := uuid.New()
	_, err := store.CreateReport(context.Background(), reports.Report{
		ReportID: id, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
		Status: reports.StatusPending, CreatedAt: old, UpdatedAt: old,
	})
	require.NoError(t, err)

	enq := &reports.FakeEnqueuer{}
	svc := reports.NewService(store, reports.FakeSummary{}, enq, true, 5*time.Minute, time.Minute)
	require.NoError(t, svc.ReconcilePending(context.Background()))
	require.GreaterOrEqual(t, enq.Calls, 1)
	found := false
	for _, p := range enq.Payloads {
		if job, ok := p.(reports.GenerateReportJob); ok && job.ReportID == id {
			found = true
			break
		}
	}
	assert.True(t, found, "expected reconcile to enqueue report %s", id)
}
