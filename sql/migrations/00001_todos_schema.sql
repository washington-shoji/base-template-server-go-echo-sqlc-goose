-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS todos;

CREATE TABLE IF NOT EXISTS todos.todos (
    todo_id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    label TEXT NOT NULL,
    completed BOOLEAN NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS todos.todos;
DROP SCHEMA IF EXISTS todos;
-- +goose StatementEnd
