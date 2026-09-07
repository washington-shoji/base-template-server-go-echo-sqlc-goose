package todos

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FakeStore is an in-memory Store for unit tests.
type FakeStore struct {
	mu    sync.Mutex
	todos map[uuid.UUID]Todo
}

func NewFakeStore() *FakeStore {
	return &FakeStore{todos: make(map[uuid.UUID]Todo)}
}

func (f *FakeStore) CreateTodo(_ context.Context, arg CreateParams) (Todo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := Todo{TodoID: arg.TodoID, CreatedAt: arg.CreatedAt, UpdatedAt: arg.UpdatedAt, Label: arg.Label, Completed: arg.Completed}
	f.todos[t.TodoID] = t
	return t, nil
}

func (f *FakeStore) UpdateTodo(_ context.Context, arg UpdateParams) (Todo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.todos[arg.TodoID]
	if !ok {
		return Todo{}, sql.ErrNoRows
	}
	t.Label, t.Completed, t.UpdatedAt = arg.Label, arg.Completed, arg.UpdatedAt
	f.todos[arg.TodoID] = t
	return t, nil
}

func (f *FakeStore) DeleteTodo(_ context.Context, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.todos[id]; !ok {
		return sql.ErrNoRows
	}
	delete(f.todos, id)
	return nil
}

func (f *FakeStore) FindTodoById(_ context.Context, id uuid.UUID) (Todo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.todos[id]
	if !ok {
		return Todo{}, sql.ErrNoRows
	}
	return t, nil
}

func (f *FakeStore) ListAllTodos(_ context.Context) ([]Todo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Todo, 0, len(f.todos))
	for _, t := range f.todos {
		out = append(out, t)
	}
	return out, nil
}

func (f *FakeStore) CountTodos(_ context.Context) (Summary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var total, completed int
	for _, t := range f.todos {
		total++
		if t.Completed {
			completed++
		}
	}
	return Summary{Total: total, Completed: completed, Outstanding: total - completed}, nil
}

func (f *FakeStore) Seed(t Todo) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}
	f.todos[t.TodoID] = t
}
