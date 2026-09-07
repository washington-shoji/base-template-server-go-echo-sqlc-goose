package postgres

import (
	"context"
	"database/sql"
	"time"

	"go-echo-server-template/internal/reports"

	"github.com/google/uuid"
)

type Store struct {
	q *Queries
}

func NewStore(db DBTX) *Store {
	return &Store{q: New(db)}
}

func (s *Store) CreateReport(ctx context.Context, r reports.Report) (reports.Report, error) {
	row, err := s.q.CreateReport(ctx, CreateReportParams{
		ReportID:            r.ReportID,
		OwnerUserID:         r.OwnerUserID,
		ReportType:          r.Type,
		Status:              r.Status,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		StartedAt:           nullTime(r.StartedAt),
		CompletedAt:         nullTime(r.CompletedAt),
		ProcessingStartedAt: nullTime(r.ProcessingStartedAt),
		FailureReason:       nullString(r.FailureReason),
		FailureCode:         nullString(r.FailureCode),
		Result:              nullString(r.Result),
	})
	if err != nil {
		return reports.Report{}, err
	}
	return mapReport(row), nil
}

func (s *Store) GetReportByID(ctx context.Context, id uuid.UUID) (reports.Report, error) {
	row, err := s.q.GetReportByID(ctx, id)
	if err != nil {
		return reports.Report{}, err
	}
	return mapReport(row), nil
}

func (s *Store) ListReportsByOwner(ctx context.Context, ownerUserID string) ([]reports.Report, error) {
	rows, err := s.q.ListReportsByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	out := make([]reports.Report, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapReport(row))
	}
	return out, nil
}

func (s *Store) ClaimExecution(ctx context.Context, id uuid.UUID, now, staleBefore time.Time) (reports.Report, error) {
	row, err := s.q.ClaimReportExecution(ctx, ClaimReportExecutionParams{
		ReportID:    id,
		NowAt:       sql.NullTime{Time: now, Valid: true},
		StaleBefore: sql.NullTime{Time: staleBefore, Valid: true},
	})
	if err != nil {
		return reports.Report{}, err
	}
	return mapReport(row), nil
}

func (s *Store) MarkPendingRetry(ctx context.Context, id uuid.UUID, at time.Time) (reports.Report, error) {
	row, err := s.q.MarkReportPendingRetry(ctx, MarkReportPendingRetryParams{
		ReportID: id,
		NowAt:    at,
	})
	if err != nil {
		return reports.Report{}, err
	}
	return mapReport(row), nil
}

func (s *Store) MarkCompleted(ctx context.Context, id uuid.UUID, result string, at time.Time) (reports.Report, error) {
	row, err := s.q.MarkReportCompleted(ctx, MarkReportCompletedParams{
		ReportID: id,
		Result:   sql.NullString{String: result, Valid: true},
		NowAt:    sql.NullTime{Time: at, Valid: true},
	})
	if err != nil {
		return reports.Report{}, err
	}
	return mapReport(row), nil
}

func (s *Store) MarkFailed(ctx context.Context, id uuid.UUID, code, reason string, at time.Time) (reports.Report, error) {
	row, err := s.q.MarkReportFailed(ctx, MarkReportFailedParams{
		ReportID:      id,
		FailureReason: sql.NullString{String: reason, Valid: reason != ""},
		FailureCode:   sql.NullString{String: code, Valid: code != ""},
		NowAt:         sql.NullTime{Time: at, Valid: true},
	})
	if err != nil {
		return reports.Report{}, err
	}
	return mapReport(row), nil
}

func (s *Store) ListPendingOlderThan(ctx context.Context, updatedBefore time.Time) ([]reports.Report, error) {
	rows, err := s.q.ListPendingReportsOlderThan(ctx, updatedBefore)
	if err != nil {
		return nil, err
	}
	out := make([]reports.Report, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapReport(row))
	}
	return out, nil
}

func mapReport(r ReportsReport) reports.Report {
	out := reports.Report{
		ReportID:    r.ReportID,
		OwnerUserID: r.OwnerUserID,
		Type:        r.ReportType,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.StartedAt.Valid {
		t := r.StartedAt.Time
		out.StartedAt = &t
	}
	if r.CompletedAt.Valid {
		t := r.CompletedAt.Time
		out.CompletedAt = &t
	}
	if r.ProcessingStartedAt.Valid {
		t := r.ProcessingStartedAt.Time
		out.ProcessingStartedAt = &t
	}
	if r.FailureReason.Valid {
		out.FailureReason = r.FailureReason.String
	}
	if r.FailureCode.Valid {
		out.FailureCode = r.FailureCode.String
	}
	if r.Result.Valid {
		out.Result = r.Result.String
	}
	return out
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
