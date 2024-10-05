package services

import (
	"context"
	"go-echo-server-template/internal/database"
	"time"

	"github.com/google/uuid"
)

type TodoServiceImpl struct {
	Context context.Context
	Query   *database.Queries
}

func NewTodoService(context context.Context, query *database.Queries) TodoService {
	return &TodoServiceImpl{
		Context: context,
		Query:   query,
	}
}

// CreateTodo implements TodoService.
func (t *TodoServiceImpl) CreateTodo(reqModel TodoParams) (database.Todo, error) {
	model := database.CreateTodoParams{
		TodoID:    uuid.New(),
		Label:     reqModel.Label,
		Completed: reqModel.Completed,
		CreatedAt: time.Now().UTC(),
	}

	result, err := t.Query.CreateTodo(t.Context, model)
	if err != nil {
		return database.Todo{}, err
	}

	return result, nil
}

// UpdateDoto implements TodoService.
func (t *TodoServiceImpl) UpdateDoto(todoId string, reqModel TodoParams) (database.Todo, error) {
	todoUUID, err := uuid.Parse(todoId)
	if err != nil {
		return database.Todo{}, err
	}
	model := database.UpdateTodoParams{
		TodoID:    todoUUID,
		Label:     reqModel.Label,
		Completed: reqModel.Completed,
		UpdatedAt: time.Now().UTC(),
	}

	result, err := t.Query.UpdateTodo(t.Context, model)
	if err != nil {
		return database.Todo{}, err
	}

	return result, nil
}

// DeleteTodo implements TodoService.
func (t *TodoServiceImpl) DeleteTodo(todoId string) error {
	todoUUID, err := uuid.Parse(todoId)
	if err != nil {
		return err
	}
	err = t.Query.DeleteTodo(t.Context, todoUUID)
	if err != nil {
		return err
	}

	return nil
}

// FindTodoById implements TodoService.
func (t *TodoServiceImpl) FindTodoById(todoId string) (database.Todo, error) {
	todoUUID, err := uuid.Parse(todoId)
	if err != nil {
		return database.Todo{}, err
	}
	result, err := t.Query.FindTodoById(t.Context, todoUUID)
	if err != nil {
		return database.Todo{}, err
	}

	return result, nil
}

// ListAllTodos implements TodoService.
func (t *TodoServiceImpl) ListAllTodos() ([]database.Todo, error) {
	result, err := t.Query.ListAllTodos(t.Context)
	if err != nil {
		return []database.Todo{}, err
	}

	return result, nil
}
