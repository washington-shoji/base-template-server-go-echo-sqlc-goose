-- name: CreateReport :one
INSERT INTO reports.reports (
    report_id, owner_user_id, report_type, status,
    created_at, updated_at, started_at, completed_at,
    processing_started_at, failure_reason, failure_code, result
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetReportByID :one
SELECT * FROM reports.reports WHERE report_id = $1 LIMIT 1;

-- name: ListReportsByOwner :many
SELECT * FROM reports.reports
WHERE owner_user_id = $1
ORDER BY created_at DESC;

-- name: ClaimReportExecution :one
UPDATE reports.reports
SET status = 'processing',
    processing_started_at = sqlc.arg(now_at),
    updated_at = sqlc.arg(now_at),
    started_at = COALESCE(started_at, sqlc.arg(now_at)),
    failure_reason = NULL,
    failure_code = NULL
WHERE report_id = sqlc.arg(report_id)
  AND (
    status = 'pending'
    OR (status = 'processing' AND (processing_started_at IS NULL OR processing_started_at < sqlc.arg(stale_before)))
  )
RETURNING *;

-- name: MarkReportPendingRetry :one
UPDATE reports.reports
SET status = 'pending',
    processing_started_at = NULL,
    updated_at = sqlc.arg(now_at),
    failure_reason = NULL,
    failure_code = NULL
WHERE report_id = sqlc.arg(report_id)
  AND status = 'processing'
RETURNING *;

-- name: MarkReportCompleted :one
UPDATE reports.reports
SET status = 'completed',
    result = sqlc.arg(result),
    completed_at = sqlc.arg(now_at),
    updated_at = sqlc.arg(now_at),
    processing_started_at = NULL,
    failure_reason = NULL,
    failure_code = NULL
WHERE report_id = sqlc.arg(report_id)
  AND status = 'processing'
RETURNING *;

-- name: MarkReportFailed :one
UPDATE reports.reports
SET status = 'failed',
    failure_reason = sqlc.arg(failure_reason),
    failure_code = sqlc.arg(failure_code),
    completed_at = sqlc.arg(now_at),
    updated_at = sqlc.arg(now_at),
    processing_started_at = NULL
WHERE report_id = sqlc.arg(report_id)
  AND status = 'processing'
RETURNING *;

-- name: ListPendingReportsOlderThan :many
SELECT * FROM reports.reports
WHERE status = 'pending'
  AND updated_at < sqlc.arg(updated_before)
ORDER BY updated_at ASC;
