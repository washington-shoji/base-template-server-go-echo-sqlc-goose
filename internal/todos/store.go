package todos

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Store is the persistence port for the todos domain (sqlc-shaped).
type Store interface {
	CreateTodo(ctx context.Context, arg CreateParams) (Todo, error)
	UpdateTodo(ctx context.Context, arg UpdateParams) (Todo, error)
	DeleteTodo(ctx context.Context, todoID uuid.UUID) error
	FindTodoById(ctx context.Context, todoID uuid.UUID) (Todo, error)
	ListAllTodos(ctx context.Context) ([]Todo, error)
	CountTodos(ctx context.Context) (Summary, error)
}

type CreateParams struct {
	TodoID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Label     string
	Completed bool
}

type UpdateParams struct {
	TodoID    uuid.UUID
	UpdatedAt time.Time
	Label     string
	Completed bool
}
