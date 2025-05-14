package testutils

import (
	"context"
	"go-echo-server-template/internal/database"

	"github.com/google/uuid"
)

// MockQueries implements database.Querier interface for testing
type MockQueries struct {
	todos map[uuid.UUID]database.Todo
}

// NewMockQueries creates a new MockQueries instance
func NewMockQueries() *MockQueries {
	return &MockQueries{
		todos: make(map[uuid.UUID]database.Todo),
	}
}

// Verify that MockQueries implements the interfaces
var (
	_ database.Querier     = (*MockQueries)(nil)
	_ database.TodoQuerier = (*MockQueries)(nil)
)

// Implementation of TodoQuerier interface
func (m *MockQueries) CreateTodo(ctx context.Context, arg database.CreateTodoParams) (database.Todo, error) {
	todo := database.Todo{
		TodoID:    arg.TodoID,
		Label:     arg.Label,
		Completed: arg.Completed,
		CreatedAt: arg.CreatedAt,
		UpdatedAt: arg.UpdatedAt,
	}
	m.todos[todo.TodoID] = todo
	return todo, nil
}

func (m *MockQueries) UpdateTodo(ctx context.Context, arg database.UpdateTodoParams) (database.Todo, error) {
	todo, exists := m.todos[arg.TodoID]
	if !exists {
		return database.Todo{}, database.ErrRecordNotFound
	}

	todo.Label = arg.Label
	todo.Completed = arg.Completed
	todo.UpdatedAt = arg.UpdatedAt

	m.todos[todo.TodoID] = todo
	return todo, nil
}

func (m *MockQueries) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	if _, exists := m.todos[todoID]; !exists {
		return database.ErrRecordNotFound
	}

	delete(m.todos, todoID)
	return nil
}

func (m *MockQueries) FindTodoById(ctx context.Context, todoID uuid.UUID) (database.Todo, error) {
	todo, exists := m.todos[todoID]
	if !exists {
		return database.Todo{}, database.ErrRecordNotFound
	}

	return todo, nil
}

func (m *MockQueries) ListAllTodos(ctx context.Context) ([]database.Todo, error) {
	todos := make([]database.Todo, 0, len(m.todos))
	for _, todo := range m.todos {
		todos = append(todos, todo)
	}
	return todos, nil
}

// Helper methods for testing
func (m *MockQueries) AddTestTodo(todo database.Todo) {
	m.todos[todo.TodoID] = todo
}

func (m *MockQueries) ClearTodos() {
	m.todos = make(map[uuid.UUID]database.Todo)
}
