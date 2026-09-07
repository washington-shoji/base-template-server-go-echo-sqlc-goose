package reports

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
)

type FakeStore struct {
	mu      sync.Mutex
	reports map[uuid.UUID]Report
}

func NewFakeStore() *FakeStore {
	return &FakeStore{reports: make(map[uuid.UUID]Report)}
}

func (f *FakeStore) CreateReport(_ context.Context, r Report) (Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reports[r.ReportID] = r
	return r, nil
}

func (f *FakeStore) GetReportByID(_ context.Context, id uuid.UUID) (Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.reports[id]
	if !ok {
		return Report{}, sql.ErrNoRows
	}
	return r, nil
}

func (f *FakeStore) ListReportsByOwner(_ context.Context, owner string) ([]Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Report
	for _, r := range f.reports {
		if r.OwnerUserID == owner {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *FakeStore) ClaimExecution(_ context.Context, id uuid.UUID, now, staleBefore time.Time) (Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.reports[id]
	if !ok {
		return Report{}, sql.ErrNoRows
	}
	eligible := r.Status == StatusPending ||
		(r.Status == StatusProcessing && (r.ProcessingStartedAt == nil || r.ProcessingStartedAt.Before(staleBefore)))
	if !eligible {
		return Report{}, sql.ErrNoRows
	}
	r.Status = StatusProcessing
	r.UpdatedAt = now
	ps := now
	r.ProcessingStartedAt = &ps
	if r.StartedAt == nil {
		r.StartedAt = &ps
	}
	r.FailureReason = ""
	r.FailureCode = ""
	f.reports[id] = r
	return r, nil
}

func (f *FakeStore) MarkPendingRetry(_ context.Context, id uuid.UUID, at time.Time) (Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.reports[id]
	if !ok {
		return Report{}, sql.ErrNoRows
	}
	if r.Status != StatusProcessing {
		return Report{}, sql.ErrNoRows
	}
	r.Status = StatusPending
	r.ProcessingStartedAt = nil
	r.UpdatedAt = at
	r.FailureReason = ""
	r.FailureCode = ""
	f.reports[id] = r
	return r, nil
}

func (f *FakeStore) MarkCompleted(_ context.Context, id uuid.UUID, result string, at time.Time) (Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.reports[id]
	if !ok {
		return Report{}, sql.ErrNoRows
	}
	if r.Status != StatusProcessing {
		return Report{}, sql.ErrNoRows
	}
	r.Status = StatusCompleted
	r.Result = result
	r.CompletedAt = &at
	r.UpdatedAt = at
	r.ProcessingStartedAt = nil
	r.FailureReason = ""
	r.FailureCode = ""
	f.reports[id] = r
	return r, nil
}

func (f *FakeStore) MarkFailed(_ context.Context, id uuid.UUID, code, reason string, at time.Time) (Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.reports[id]
	if !ok {
		return Report{}, sql.ErrNoRows
	}
	if r.Status != StatusProcessing {
		return Report{}, sql.ErrNoRows
	}
	r.Status = StatusFailed
	r.FailureCode = code
	r.FailureReason = reason
	r.CompletedAt = &at
	r.UpdatedAt = at
	r.ProcessingStartedAt = nil
	f.reports[id] = r
	return r, nil
}

func (f *FakeStore) ListPendingOlderThan(_ context.Context, updatedBefore time.Time) ([]Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Report
	for _, r := range f.reports {
		if r.Status == StatusPending && r.UpdatedAt.Before(updatedBefore) {
			out = append(out, r)
		}
	}
	return out, nil
}

// Seed inserts or replaces a report without going through CreateReport (tests).
func (f *FakeStore) Seed(r Report) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reports[r.ReportID] = r
}

type FakeSummary struct {
	Value TodoSummary
	Err   error
}

func (f FakeSummary) Summary(context.Context) (TodoSummary, error) {
	return f.Value, f.Err
}

type FakeEnqueuer struct {
	Err      error
	Calls    int
	LastJob  string
	LastPay  any
	Payloads []any
}

func (f *FakeEnqueuer) Enqueue(_ context.Context, jobType string, payload any) error {
	f.Calls++
	f.LastJob = jobType
	f.LastPay = payload
	f.Payloads = append(f.Payloads, payload)
	return f.Err
}
