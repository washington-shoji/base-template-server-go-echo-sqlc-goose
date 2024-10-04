-- name: CreateTodo :one

INSERT INTO todos (todo_id, created_at, updated_at, label, completed)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
