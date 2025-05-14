package database

import (
	"context"

	"github.com/google/uuid"
)

// Querier defines the interface that both the real database and mock implementations must satisfy
type Querier interface {
	CreateTodo(ctx context.Context, arg CreateTodoParams) (Todo, error)
	UpdateTodo(ctx context.Context, arg UpdateTodoParams) (Todo, error)
	DeleteTodo(ctx context.Context, todoID uuid.UUID) error
	FindTodoById(ctx context.Context, todoID uuid.UUID) (Todo, error)
	ListAllTodos(ctx context.Context) ([]Todo, error)
}

// Verify that Queries implements Querier
var _ Querier = (*Queries)(nil)
