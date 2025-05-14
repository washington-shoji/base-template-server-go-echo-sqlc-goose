package services

import (
	"context"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/internal/logger"
	"time"

	"github.com/google/uuid"
)

type TodoServiceImpl struct {
	Context context.Context
	Query   *database.Queries
	log     *logger.Logger
}

func NewTodoService(context context.Context, query *database.Queries) TodoService {
	return &TodoServiceImpl{
		Context: context,
		Query:   query,
		log:     logger.WithContext(context, "todo_service"),
	}
}

// CreateTodo implements TodoService.
func (t *TodoServiceImpl) CreateTodo(reqModel TodoParams) (database.Todo, error) {
	// Log the start of todo creation with input parameters
	t.log.Debug("Creating new todo", map[string]interface{}{
		"label":     reqModel.Label,
		"completed": reqModel.Completed,
	})

	model := database.CreateTodoParams{
		TodoID:    uuid.New(),
		Label:     reqModel.Label,
		Completed: reqModel.Completed,
		CreatedAt: time.Now().UTC(),
	}

	result, err := t.Query.CreateTodo(t.Context, model)
	if err != nil {
		// Log database error with todo ID for tracking
		t.log.Error("Failed to create todo", err, map[string]interface{}{
			"todo_id": model.TodoID,
		})
		return database.Todo{}, err
	}

	// Log successful todo creation with the generated ID
	t.log.Debug("Todo created successfully", map[string]interface{}{
		"todo_id": result.TodoID,
	})
	return result, nil
}

// UpdateDoto implements TodoService.
func (t *TodoServiceImpl) UpdateDoto(todoId string, reqModel TodoParams) (database.Todo, error) {
	// Log the start of todo update with input parameters
	t.log.Debug("Updating todo", map[string]interface{}{
		"todo_id":   todoId,
		"label":     reqModel.Label,
		"completed": reqModel.Completed,
	})

	todoUUID, err := uuid.Parse(todoId)
	if err != nil {
		// Log invalid UUID format error
		t.log.Error("Invalid todo ID format", err, map[string]interface{}{
			"todo_id": todoId,
		})
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
		// Log database error during update
		t.log.Error("Failed to update todo", err, map[string]interface{}{
			"todo_id": todoUUID,
		})
		return database.Todo{}, err
	}

	// Log successful todo update
	t.log.Debug("Todo updated successfully", map[string]interface{}{
		"todo_id": result.TodoID,
	})
	return result, nil
}

// DeleteTodo implements TodoService.
func (t *TodoServiceImpl) DeleteTodo(todoId string) error {
	// Log the start of todo deletion
	t.log.Debug("Deleting todo", map[string]interface{}{
		"todo_id": todoId,
	})

	todoUUID, err := uuid.Parse(todoId)
	if err != nil {
		// Log invalid UUID format error
		t.log.Error("Invalid todo ID format", err, map[string]interface{}{
			"todo_id": todoId,
		})
		return err
	}

	err = t.Query.DeleteTodo(t.Context, todoUUID)
	if err != nil {
		// Log database error during deletion
		t.log.Error("Failed to delete todo", err, map[string]interface{}{
			"todo_id": todoUUID,
		})
		return err
	}

	// Log successful todo deletion
	t.log.Debug("Todo deleted successfully", map[string]interface{}{
		"todo_id": todoUUID,
	})
	return nil
}

// FindTodoById implements TodoService.
func (t *TodoServiceImpl) FindTodoById(todoId string) (database.Todo, error) {
	// Log the start of todo search
	t.log.Debug("Finding todo by ID", map[string]interface{}{
		"todo_id": todoId,
	})

	todoUUID, err := uuid.Parse(todoId)
	if err != nil {
		// Log invalid UUID format error
		t.log.Error("Invalid todo ID format", err, map[string]interface{}{
			"todo_id": todoId,
		})
		return database.Todo{}, err
	}

	result, err := t.Query.FindTodoById(t.Context, todoUUID)
	if err != nil {
		// Log database error during search
		t.log.Error("Failed to find todo", err, map[string]interface{}{
			"todo_id": todoUUID,
		})
		return database.Todo{}, err
	}

	// Log successful todo retrieval
	t.log.Debug("Todo found successfully", map[string]interface{}{
		"todo_id": result.TodoID,
	})
	return result, nil
}

// ListAllTodos implements TodoService.
func (t *TodoServiceImpl) ListAllTodos() ([]database.Todo, error) {
	// Log the start of listing all todos
	t.log.Debug("Listing all todos", map[string]interface{}{})

	result, err := t.Query.ListAllTodos(t.Context)
	if err != nil {
		// Log database error during listing
		t.log.Error("Failed to list todos", err, nil)
		return []database.Todo{}, err
	}

	// Log successful listing with count of todos
	t.log.Debug("Todos listed successfully", map[string]interface{}{
		"count": len(result),
	})
	return result, nil
}
