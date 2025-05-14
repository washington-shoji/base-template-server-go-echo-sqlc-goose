package testutils

import (
	"context"
	"go-echo-server-template/internal/database"
	"time"

	"github.com/google/uuid"
)

// MockQueries implements database.Queries for testing
type MockQueries struct {
	todos map[uuid.UUID]database.Todo
}

// Verify that MockQueries implements database.Querier
var _ database.Querier = (*MockQueries)(nil)

// NewMockQueries creates a new mock database
func NewMockQueries() *MockQueries {
	return &MockQueries{
		todos: make(map[uuid.UUID]database.Todo),
	}
}

// CreateTodo implements database.Queries
func (m *MockQueries) CreateTodo(ctx context.Context, arg database.CreateTodoParams) (database.Todo, error) {
	todo := database.Todo{
		TodoID:    arg.TodoID,
		Label:     arg.Label,
		Completed: arg.Completed,
		CreatedAt: arg.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}
	m.todos[arg.TodoID] = todo
	return todo, nil
}

// UpdateTodo implements database.Queries
func (m *MockQueries) UpdateTodo(ctx context.Context, arg database.UpdateTodoParams) (database.Todo, error) {
	if _, exists := m.todos[arg.TodoID]; exists {
		// Add a small delay to ensure UpdatedAt is after the original time
		time.Sleep(1 * time.Millisecond)
		todo := database.Todo{
			TodoID:    arg.TodoID,
			Label:     arg.Label,
			Completed: arg.Completed,
			CreatedAt: m.todos[arg.TodoID].CreatedAt,
			UpdatedAt: time.Now().UTC(),
		}
		m.todos[arg.TodoID] = todo
		return todo, nil
	}
	return database.Todo{}, database.ErrRecordNotFound
}

// DeleteTodo implements database.Queries
func (m *MockQueries) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	if _, exists := m.todos[todoID]; exists {
		delete(m.todos, todoID)
		return nil
	}
	return database.ErrRecordNotFound
}

// FindTodoById implements database.Queries
func (m *MockQueries) FindTodoById(ctx context.Context, todoID uuid.UUID) (database.Todo, error) {
	if todo, exists := m.todos[todoID]; exists {
		return todo, nil
	}
	return database.Todo{}, database.ErrRecordNotFound
}

// ListAllTodos implements database.Queries
func (m *MockQueries) ListAllTodos(ctx context.Context) ([]database.Todo, error) {
	todos := make([]database.Todo, 0, len(m.todos))
	for _, todo := range m.todos {
		todos = append(todos, todo)
	}
	return todos, nil
}

// AddTestTodo adds a todo for testing
func (m *MockQueries) AddTestTodo(todo database.Todo) {
	m.todos[todo.TodoID] = todo
}

// ClearTodos clears all todos for testing
func (m *MockQueries) ClearTodos() {
	m.todos = make(map[uuid.UUID]database.Todo)
}
