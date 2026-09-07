package reports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	CreateReport(ctx context.Context, r Report) (Report, error)
	GetReportByID(ctx context.Context, id uuid.UUID) (Report, error)
	ListReportsByOwner(ctx context.Context, ownerUserID string) ([]Report, error)
	// ClaimExecution moves pending (or stale processing) → processing.
	// Returns sql.ErrNoRows when another worker owns a fresh lease.
	ClaimExecution(ctx context.Context, id uuid.UUID, now, staleBefore time.Time) (Report, error)
	MarkPendingRetry(ctx context.Context, id uuid.UUID, at time.Time) (Report, error)
	MarkCompleted(ctx context.Context, id uuid.UUID, result string, at time.Time) (Report, error)
	MarkFailed(ctx context.Context, id uuid.UUID, code, reason string, at time.Time) (Report, error)
	ListPendingOlderThan(ctx context.Context, updatedBefore time.Time) ([]Report, error)
}
