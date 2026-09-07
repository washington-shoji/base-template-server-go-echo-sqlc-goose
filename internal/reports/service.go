package reports

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/authz"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/jobs"
	"go-echo-server-template/internal/platform/logging"

	"github.com/google/uuid"
)

type Service struct {
	store            Store
	summaries        TodoSummaryReader
	enqueue          Enqueuer
	authDisabled     bool
	processingLease  time.Duration
	reconcileAfter   time.Duration
}

func NewService(
	store Store,
	summaries TodoSummaryReader,
	enqueue Enqueuer,
	authDisabled bool,
	processingLease, reconcileAfter time.Duration,
) *Service {
	if processingLease <= 0 {
		processingLease = 5 * time.Minute
	}
	if reconcileAfter <= 0 {
		reconcileAfter = time.Minute
	}
	return &Service{
		store:           store,
		summaries:       summaries,
		enqueue:         enqueue,
		authDisabled:    authDisabled,
		processingLease: processingLease,
		reconcileAfter:  reconcileAfter,
	}
}

func (s *Service) principal(ctx context.Context) (authn.Principal, error) {
	p, _ := authn.FromContext(ctx)
	if err := authz.RequireAuthenticated(p, s.authDisabled); err != nil {
		return authn.Principal{}, fmt.Errorf("%w: authentication required", httpx.ErrPermissionDenied)
	}
	if !s.authDisabled && p.UserID == "" {
		return authn.Principal{}, fmt.Errorf("%w: authentication required", httpx.ErrPermissionDenied)
	}
	return p, nil
}

func (s *Service) assertOwner(p authn.Principal, r Report) error {
	if r.OwnerUserID != p.UserID {
		return fmt.Errorf("%w: not your report", httpx.ErrPermissionDenied)
	}
	return nil
}

func supportedType(t string) bool {
	return t == TypeTodoSummary
}

// CreateReport inserts a pending report and enqueues generation.
// On enqueue failure the row stays pending for ReconcilePending (no MarkFailed).
func (s *Service) CreateReport(ctx context.Context, req CreateRequest) (Report, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Report{}, err
	}
	if req.Type == "" {
		return Report{}, fmt.Errorf("%w: type is required", httpx.ErrValidation)
	}
	if !supportedType(req.Type) {
		return Report{}, fmt.Errorf("%w: unsupported report type", httpx.ErrValidation)
	}

	now := time.Now().UTC()
	r := Report{
		ReportID:    uuid.New(),
		OwnerUserID: p.UserID,
		Type:        req.Type,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created, err := s.store.CreateReport(ctx, r)
	if err != nil {
		logging.WithContext(ctx, "reports").Error("create report", err, nil)
		return Report{}, fmt.Errorf("%w: create report", httpx.ErrInternal)
	}

	if err := s.enqueue.Enqueue(ctx, JobGenerate, GenerateReportJob{ReportID: created.ReportID}); err != nil {
		logging.WithContext(ctx, "reports").Error("enqueue generate; leaving pending for reconcile", err, map[string]interface{}{
			"report_id": created.ReportID,
		})
		return created, nil
	}
	return created, nil
}

func (s *Service) GetReport(ctx context.Context, id string) (Report, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Report{}, err
	}
	rid, err := uuid.Parse(id)
	if err != nil {
		return Report{}, fmt.Errorf("%w: invalid report id", httpx.ErrBadRequest)
	}
	r, err := s.store.GetReportByID(ctx, rid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, fmt.Errorf("%w: report not found", httpx.ErrNotFound)
		}
		return Report{}, fmt.Errorf("%w: get report", httpx.ErrInternal)
	}
	if err := s.assertOwner(p, r); err != nil {
		return Report{}, err
	}
	return r, nil
}

func (s *Service) ListReports(ctx context.Context) ([]Report, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.store.ListReportsByOwner(ctx, p.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: list reports", httpx.ErrInternal)
	}
	return list, nil
}

// Generate runs report generation. Safe under at-least-once job delivery.
// Retryable failures return RetryableError after processing → pending (unless last job attempt).
func (s *Service) Generate(ctx context.Context, reportID uuid.UUID) error {
	log := logging.WithContext(ctx, "reports")
	r, err := s.store.GetReportByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return Retryable(err)
	}

	switch r.Status {
	case StatusCompleted, StatusFailed:
		return nil
	case StatusPending, StatusProcessing:
		// claim below
	default:
		return nil
	}

	now := time.Now().UTC()
	staleBefore := now.Add(-s.processingLease)
	claimed, err := s.store.ClaimExecution(ctx, reportID, now, staleBefore)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Concurrent owner or not eligible.
			return nil
		}
		return Retryable(err)
	}
	_ = claimed

	if !supportedType(r.Type) {
		_, _ = s.store.MarkFailed(ctx, reportID, FailureUnsupportedType,
			SafeFailureMessage(FailureUnsupportedType, ""), time.Now().UTC())
		return nil
	}

	summary, err := s.summaries.Summary(ctx)
	if err != nil {
		log.Error("todo summary capability failed", err, map[string]interface{}{"report_id": reportID})
		return s.handleRetryable(ctx, reportID, err)
	}

	result := fmt.Sprintf("Total todos: %d\nCompleted: %d\nOutstanding: %d",
		summary.Total, summary.Completed, summary.Outstanding)

	if _, err := s.store.MarkCompleted(ctx, reportID, result, time.Now().UTC()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return Retryable(err)
	}
	return nil
}

func (s *Service) handleRetryable(ctx context.Context, reportID uuid.UUID, cause error) error {
	if info, ok := jobs.AttemptFromContext(ctx); ok && info.Attempt >= info.MaxAttempts {
		_, _ = s.store.MarkFailed(ctx, reportID, FailureRetryExhausted,
			SafeFailureMessage(FailureRetryExhausted, ""), time.Now().UTC())
		return nil
	}
	if _, err := s.store.MarkPendingRetry(ctx, reportID, time.Now().UTC()); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Retryable(err)
	}
	return Retryable(cause)
}

// ReconcilePending re-enqueues pending reports that have not progressed.
// Duplicates are safe: Generate claim/idempotency handles them.
func (s *Service) ReconcilePending(ctx context.Context) error {
	before := time.Now().UTC().Add(-s.reconcileAfter)
	list, err := s.store.ListPendingOlderThan(ctx, before)
	if err != nil {
		return err
	}
	log := logging.WithContext(ctx, "reports")
	for _, r := range list {
		if err := s.enqueue.Enqueue(ctx, JobGenerate, GenerateReportJob{ReportID: r.ReportID}); err != nil {
			log.Error("reconcile enqueue", err, map[string]interface{}{"report_id": r.ReportID})
			continue
		}
	}
	return nil
}
