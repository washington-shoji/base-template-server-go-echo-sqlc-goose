package database

import (
	"context"

	"github.com/google/uuid"
)

// TodoQuerier defines the interface for todo-related database operations
type TodoQuerier interface {
	CreateTodo(ctx context.Context, arg CreateTodoParams) (Todo, error)
	UpdateTodo(ctx context.Context, arg UpdateTodoParams) (Todo, error)
	DeleteTodo(ctx context.Context, todoID uuid.UUID) error
	FindTodoById(ctx context.Context, todoID uuid.UUID) (Todo, error)
	ListAllTodos(ctx context.Context) ([]Todo, error)
}

// Querier combines all domain-specific queriers
type Querier interface {
	TodoQuerier
	// When you add AddressQuerier, it will be added here like:
	// AddressQuerier
}

// Verify that Queries implements Querier
var _ Querier = (*Queries)(nil)
