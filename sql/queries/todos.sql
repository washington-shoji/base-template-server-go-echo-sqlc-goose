-- name: CreateTodo :one

INSERT INTO todos (todo_id, created_at, updated_at, label, completed)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: UpdateTodo :one

UPDATE todos
SET
    updated_at = $2, 
    label = $3, 
    completed = $4
WHERE todo_id = $1    
RETURNING *;


-- name: DeleteTodo :exec

DELETE FROM todos
WHERE todo_id = $1;


-- name: FindTodoById :one

SELECT * FROM todos
WHERE todo_id = $1 LIMIT 1;


-- name: ListAllTodos :many

SELECT * FROM todos
ORDER BY created_at DESC;