package reports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"

	TypeTodoSummary = "todo-summary"

	JobGenerate = "reports.generate"

	FailureUnsupportedType = "UNSUPPORTED_TYPE"
	FailureRetryExhausted  = "RETRY_EXHAUSTED"
	FailureInternal        = "INTERNAL"
)

// Report is owned by a user (Principal.UserID) and generated asynchronously.
// Queue attempt/lock fields are never exposed on this type.
type Report struct {
	ReportID            uuid.UUID  `json:"report_id"`
	OwnerUserID         string     `json:"owner_user_id"`
	Type                string     `json:"type"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	ProcessingStartedAt *time.Time `json:"-"`
	FailureReason       string     `json:"failure_reason,omitempty"` // safe user-facing message
	FailureCode         string     `json:"failure_code,omitempty"`
	Result              string     `json:"result,omitempty"`
}

type CreateRequest struct {
	Type string `json:"type"`
}

// GenerateReportJob is the typed job payload for reports.generate.
type GenerateReportJob struct {
	ReportID uuid.UUID `json:"report_id"`
}

// TodoSummary is the cross-domain read model (global counts; Todos are not user-owned).
type TodoSummary struct {
	Total       int `json:"total"`
	Completed   int `json:"completed"`
	Outstanding int `json:"outstanding"`
}

// TodoSummaryReader is a consumer-owned capability; implemented by Todos via app adapter.
type TodoSummaryReader interface {
	Summary(ctx context.Context) (TodoSummary, error)
}

// Enqueuer abstracts job enqueue so Reports does not depend on jobs SQL.
type Enqueuer interface {
	Enqueue(ctx context.Context, jobType string, payload any) error
}

// RetryableError marks an error as eligible for another Generate attempt.
// Prefer returning this after MarkPendingRetry so the job queue backs off.
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	if e.Err == nil {
		return "retryable"
	}
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error { return e.Err }

func Retryable(err error) error {
	if err == nil {
		return nil
	}
	var r *RetryableError
	if errors.As(err, &r) {
		return err
	}
	return &RetryableError{Err: err}
}

func IsRetryable(err error) bool {
	var r *RetryableError
	return errors.As(err, &r)
}

func SafeFailureMessage(code string, detail string) string {
	switch code {
	case FailureUnsupportedType:
		return "unsupported report type"
	case FailureRetryExhausted:
		return "report generation failed after retries"
	case FailureInternal:
		if detail != "" {
			return detail
		}
		return "report generation failed"
	default:
		if detail != "" {
			return detail
		}
		return fmt.Sprintf("report failed (%s)", code)
	}
}
