package reports_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/jobs"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/reports"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { _ = logging.Initialize("test", "ERROR") }

func userCtx(uid string) context.Context {
	return authn.WithPrincipal(context.Background(), authn.Principal{
		UserID: uid, Roles: []string{"user"}, AuthMethod: "test",
	})
}

func newSvc(store reports.Store, sum reports.TodoSummaryReader, enq reports.Enqueuer, authDisabled bool) *reports.Service {
	return reports.NewService(store, sum, enq, authDisabled, time.Minute, time.Minute)
}

func TestCreateAndGenerate(t *testing.T) {
	store := reports.NewFakeStore()
	enq := &reports.FakeEnqueuer{}
	sum := reports.FakeSummary{Value: reports.TodoSummary{Total: 3, Completed: 1, Outstanding: 2}}
	svc := newSvc(store, sum, enq, false)

	r, err := svc.CreateReport(userCtx("u1"), reports.CreateRequest{Type: reports.TypeTodoSummary})
	require.NoError(t, err)
	assert.Equal(t, reports.StatusPending, r.Status)
	assert.Equal(t, "u1", r.OwnerUserID)
	assert.Equal(t, 1, enq.Calls)

	require.NoError(t, svc.Generate(userCtx("system"), r.ReportID))
	got, err := svc.GetReport(userCtx("u1"), r.ReportID.String())
	require.NoError(t, err)
	assert.Equal(t, reports.StatusCompleted, got.Status)
	assert.Contains(t, got.Result, "Total todos: 3")
	assert.Contains(t, got.Result, "Completed: 1")
}

func TestUnsupportedType(t *testing.T) {
	svc := newSvc(reports.NewFakeStore(), reports.FakeSummary{}, &reports.FakeEnqueuer{}, false)
	_, err := svc.CreateReport(userCtx("u1"), reports.CreateRequest{Type: "nope"})
	assert.True(t, errors.Is(err, httpx.ErrValidation))
}

func TestOwnership(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{Value: reports.TodoSummary{}}, &reports.FakeEnqueuer{}, false)
	r, err := svc.CreateReport(userCtx("owner"), reports.CreateRequest{Type: reports.TypeTodoSummary})
	require.NoError(t, err)

	_, err = svc.GetReport(userCtx("other"), r.ReportID.String())
	assert.True(t, errors.Is(err, httpx.ErrPermissionDenied))

	list, err := svc.ListReports(userCtx("other"))
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestIdempotentGenerate(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{Value: reports.TodoSummary{Total: 1}}, &reports.FakeEnqueuer{}, true)
	r, err := svc.CreateReport(userCtx("anonymous"), reports.CreateRequest{Type: reports.TypeTodoSummary})
	require.NoError(t, err)
	require.NoError(t, svc.Generate(context.Background(), r.ReportID))
	require.NoError(t, svc.Generate(context.Background(), r.ReportID))
	got, _ := store.GetReportByID(context.Background(), r.ReportID)
	assert.Equal(t, reports.StatusCompleted, got.Status)
}

func TestEnqueueFailureLeavesPending(t *testing.T) {
	store := reports.NewFakeStore()
	enq := &reports.FakeEnqueuer{Err: fmt.Errorf("queue down")}
	svc := newSvc(store, reports.FakeSummary{}, enq, false)
	r, err := svc.CreateReport(userCtx("u1"), reports.CreateRequest{Type: reports.TypeTodoSummary})
	require.NoError(t, err)
	assert.Equal(t, reports.StatusPending, r.Status)
	assert.Empty(t, r.FailureCode)
}

func TestRetryableSummaryErrorReturnsPending(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{Err: errors.New("db blip")}, &reports.FakeEnqueuer{}, false)
	r, err := svc.CreateReport(userCtx("u1"), reports.CreateRequest{Type: reports.TypeTodoSummary})
	require.NoError(t, err)
	err = svc.Generate(userCtx("system"), r.ReportID)
	require.Error(t, err)
	assert.True(t, reports.IsRetryable(err))
	got, _ := store.GetReportByID(context.Background(), r.ReportID)
	assert.Equal(t, reports.StatusPending, got.Status)
	assert.Nil(t, got.ProcessingStartedAt)
}

func TestLastAttemptMarksFailed(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{Err: errors.New("db blip")}, &reports.FakeEnqueuer{}, false)
	r, err := svc.CreateReport(userCtx("u1"), reports.CreateRequest{Type: reports.TypeTodoSummary})
	require.NoError(t, err)

	ctx := jobs.ContextWithAttempt(userCtx("system"), jobs.AttemptInfo{Attempt: 5, MaxAttempts: 5})
	require.NoError(t, svc.Generate(ctx, r.ReportID))
	got, _ := store.GetReportByID(context.Background(), r.ReportID)
	assert.Equal(t, reports.StatusFailed, got.Status)
	assert.Equal(t, reports.FailureRetryExhausted, got.FailureCode)
}

func TestClaimRaceSecondNoOp(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{Value: reports.TodoSummary{Total: 1}}, &reports.FakeEnqueuer{}, true)
	r, err := svc.CreateReport(userCtx("anonymous"), reports.CreateRequest{Type: reports.TypeTodoSummary})
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = store.ClaimExecution(context.Background(), r.ReportID, now, now.Add(-time.Minute))
	require.NoError(t, err)

	// Fresh lease held; Generate should no-op without completing.
	require.NoError(t, svc.Generate(context.Background(), r.ReportID))
	got, _ := store.GetReportByID(context.Background(), r.ReportID)
	assert.Equal(t, reports.StatusProcessing, got.Status)
	assert.Empty(t, got.Result)
}

func TestStaleProcessingReclaim(t *testing.T) {
	store := reports.NewFakeStore()
	svc := reports.NewService(store, reports.FakeSummary{Value: reports.TodoSummary{Total: 2}}, &reports.FakeEnqueuer{}, true, time.Minute, time.Minute)
	id := uuid.New()
	stale := time.Now().UTC().Add(-2 * time.Minute)
	store.Seed(reports.Report{
		ReportID:            id,
		OwnerUserID:         "anonymous",
		Type:                reports.TypeTodoSummary,
		Status:              reports.StatusProcessing,
		CreatedAt:           stale,
		UpdatedAt:           stale,
		StartedAt:           &stale,
		ProcessingStartedAt: &stale,
	})
	require.NoError(t, svc.Generate(context.Background(), id))
	got, _ := store.GetReportByID(context.Background(), id)
	assert.Equal(t, reports.StatusCompleted, got.Status)
}

func TestTerminalUnsupportedType(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{}, &reports.FakeEnqueuer{}, true)
	id := uuid.New()
	now := time.Now().UTC()
	store.Seed(reports.Report{
		ReportID:    id,
		OwnerUserID: "anonymous",
		Type:        "weird",
		Status:      reports.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	require.NoError(t, svc.Generate(context.Background(), id))
	got, _ := store.GetReportByID(context.Background(), id)
	assert.Equal(t, reports.StatusFailed, got.Status)
	assert.Equal(t, reports.FailureUnsupportedType, got.FailureCode)
}

func TestCompletedAndFailedNoOp(t *testing.T) {
	store := reports.NewFakeStore()
	svc := newSvc(store, reports.FakeSummary{Value: reports.TodoSummary{Total: 99}}, &reports.FakeEnqueuer{}, true)
	now := time.Now().UTC()
	doneID := uuid.New()
	failID := uuid.New()
	store.Seed(reports.Report{
		ReportID: doneID, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
		Status: reports.StatusCompleted, Result: "old", CreatedAt: now, UpdatedAt: now,
	})
	store.Seed(reports.Report{
		ReportID: failID, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
		Status: reports.StatusFailed, FailureCode: reports.FailureInternal, CreatedAt: now, UpdatedAt: now,
	})
	require.NoError(t, svc.Generate(context.Background(), doneID))
	require.NoError(t, svc.Generate(context.Background(), failID))
	done, _ := store.GetReportByID(context.Background(), doneID)
	fail, _ := store.GetReportByID(context.Background(), failID)
	assert.Equal(t, "old", done.Result)
	assert.Equal(t, reports.StatusFailed, fail.Status)
}

func TestReconcilePendingEnqueues(t *testing.T) {
	store := reports.NewFakeStore()
	enq := &reports.FakeEnqueuer{}
	svc := reports.NewService(store, reports.FakeSummary{}, enq, true, time.Minute, time.Minute)
	old := time.Now().UTC().Add(-2 * time.Minute)
	fresh := time.Now().UTC()
	oldID := uuid.New()
	freshID := uuid.New()
	store.Seed(reports.Report{
		ReportID: oldID, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
		Status: reports.StatusPending, CreatedAt: old, UpdatedAt: old,
	})
	store.Seed(reports.Report{
		ReportID: freshID, OwnerUserID: "anonymous", Type: reports.TypeTodoSummary,
		Status: reports.StatusPending, CreatedAt: fresh, UpdatedAt: fresh,
	})
	require.NoError(t, svc.ReconcilePending(context.Background()))
	assert.Equal(t, 1, enq.Calls)
	pay, ok := enq.LastPay.(reports.GenerateReportJob)
	require.True(t, ok)
	assert.Equal(t, oldID, pay.ReportID)
}

func TestGenerateMissingReport(t *testing.T) {
	svc := newSvc(reports.NewFakeStore(), reports.FakeSummary{}, &reports.FakeEnqueuer{}, true)
	require.NoError(t, svc.Generate(context.Background(), uuid.New()))
}
