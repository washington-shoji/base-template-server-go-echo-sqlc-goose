package postgres

import (
	"context"

	"go-echo-server-template/internal/todos"

	"github.com/google/uuid"
)

// Store adapts sqlc Queries to the todos.Store port.
type Store struct {
	q *Queries
}

func NewStore(db DBTX) *Store {
	return &Store{q: New(db)}
}

func (s *Store) CreateTodo(ctx context.Context, arg todos.CreateParams) (todos.Todo, error) {
	row, err := s.q.CreateTodo(ctx, CreateTodoParams{
		TodoID: arg.TodoID, CreatedAt: arg.CreatedAt, UpdatedAt: arg.UpdatedAt,
		Label: arg.Label, Completed: arg.Completed,
	})
	if err != nil {
		return todos.Todo{}, err
	}
	return mapTodo(row), nil
}

func (s *Store) UpdateTodo(ctx context.Context, arg todos.UpdateParams) (todos.Todo, error) {
	row, err := s.q.UpdateTodo(ctx, UpdateTodoParams{
		TodoID: arg.TodoID, UpdatedAt: arg.UpdatedAt, Label: arg.Label, Completed: arg.Completed,
	})
	if err != nil {
		return todos.Todo{}, err
	}
	return mapTodo(row), nil
}

func (s *Store) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	return s.q.DeleteTodo(ctx, todoID)
}

func (s *Store) FindTodoById(ctx context.Context, todoID uuid.UUID) (todos.Todo, error) {
	row, err := s.q.FindTodoById(ctx, todoID)
	if err != nil {
		return todos.Todo{}, err
	}
	return mapTodo(row), nil
}

func (s *Store) ListAllTodos(ctx context.Context) ([]todos.Todo, error) {
	rows, err := s.q.ListAllTodos(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]todos.Todo, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapTodo(r))
	}
	return out, nil
}

func (s *Store) CountTodos(ctx context.Context) (todos.Summary, error) {
	row, err := s.q.CountTodos(ctx)
	if err != nil {
		return todos.Summary{}, err
	}
	return todos.Summary{
		Total:       int(row.Total),
		Completed:   int(row.Completed),
		Outstanding: int(row.Total - row.Completed),
	}, nil
}

func mapTodo(r TodosTodo) todos.Todo {
	return todos.Todo{
		TodoID: r.TodoID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		Label: r.Label, Completed: r.Completed,
	}
}
