-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS reports;

CREATE TABLE IF NOT EXISTS reports.reports (
    report_id UUID PRIMARY KEY,
    owner_user_id TEXT NOT NULL,
    report_type TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failure_reason TEXT,
    result TEXT
);

CREATE INDEX IF NOT EXISTS idx_reports_owner ON reports.reports (owner_user_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reports.reports;
DROP SCHEMA IF EXISTS reports;
-- +goose StatementEnd
