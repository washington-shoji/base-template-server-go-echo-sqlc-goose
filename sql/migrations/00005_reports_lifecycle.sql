-- +goose Up
-- +goose StatementBegin
ALTER TABLE reports.reports
    ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failure_code TEXT;

CREATE INDEX IF NOT EXISTS idx_reports_pending_updated
    ON reports.reports (updated_at)
    WHERE status = 'pending';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS reports.idx_reports_pending_updated;
ALTER TABLE reports.reports
    DROP COLUMN IF EXISTS processing_started_at,
    DROP COLUMN IF EXISTS failure_code;
-- +goose StatementEnd
