-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    user_id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    roles TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth.sessions (
    session_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES auth.users(user_id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth.api_tokens (
    token_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES auth.users(user_id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON auth.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_api_tokens_hash ON auth.api_tokens(token_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth.api_tokens;
DROP TABLE IF EXISTS auth.sessions;
DROP TABLE IF EXISTS auth.users;
DROP SCHEMA IF EXISTS auth;
-- +goose StatementEnd
