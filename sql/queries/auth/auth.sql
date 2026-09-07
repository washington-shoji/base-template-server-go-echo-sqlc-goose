-- name: CreateUser :one
INSERT INTO auth.users (user_id, email, password_hash, roles, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM auth.users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM auth.users WHERE user_id = $1 LIMIT 1;

-- name: CreateSession :one
INSERT INTO auth.sessions (session_id, user_id, expires_at, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSession :one
SELECT * FROM auth.sessions WHERE session_id = $1 LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM auth.sessions WHERE session_id = $1;

-- name: CreateAPIToken :one
INSERT INTO auth.api_tokens (token_id, user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAPITokenByHash :one
SELECT * FROM auth.api_tokens WHERE token_hash = $1 LIMIT 1;

-- name: DeleteAPIToken :exec
DELETE FROM auth.api_tokens WHERE token_id = $1;
