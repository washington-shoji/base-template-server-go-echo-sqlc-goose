package services

import (
	"context"
	"database/sql"
	"go-echo-server-template/internal/database"
	appErrors "go-echo-server-template/internal/errors"
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

	// Validate input
	if reqModel.Label == "" {
		return database.Todo{}, appErrors.NewValidationError("Label is required", map[string]interface{}{
			"field": "label",
		})
	}

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
		return database.Todo{}, appErrors.NewInternalServerError(err, "Failed to create todo")
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

	// Validate input
	if reqModel.Label == "" {
		return database.Todo{}, appErrors.NewValidationError("Label is required", map[string]interface{}{
			"field": "label",
		})
	}

	todoUUID, err := uuid.Parse(todoId)
	if err != nil {
		return database.Todo{}, appErrors.NewBadRequestError("Invalid todo ID format", map[string]interface{}{
			"todo_id": todoId,
		})
	}

	model := database.UpdateTodoParams{
		TodoID:    todoUUID,
		Label:     reqModel.Label,
		Completed: reqModel.Completed,
		UpdatedAt: time.Now().UTC(),
	}

	result, err := t.Query.UpdateTodo(t.Context, model)
	if err != nil {
		if err == sql.ErrNoRows {
			return database.Todo{}, appErrors.NewNotFoundError("Todo not found", map[string]interface{}{
				"todo_id": todoId,
			})
		}
		t.log.Error("Failed to update todo", err, map[string]interface{}{
			"todo_id": todoUUID,
		})
		return database.Todo{}, appErrors.NewInternalServerError(err, "Failed to update todo")
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
		return appErrors.NewBadRequestError("Invalid todo ID format", map[string]interface{}{
			"todo_id": todoId,
		})
	}

	err = t.Query.DeleteTodo(t.Context, todoUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return appErrors.NewNotFoundError("Todo not found", map[string]interface{}{
				"todo_id": todoId,
			})
		}
		t.log.Error("Failed to delete todo", err, map[string]interface{}{
			"todo_id": todoUUID,
		})
		return appErrors.NewInternalServerError(err, "Failed to delete todo")
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
		return database.Todo{}, appErrors.NewBadRequestError("Invalid todo ID format", map[string]interface{}{
			"todo_id": todoId,
		})
	}

	result, err := t.Query.FindTodoById(t.Context, todoUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return database.Todo{}, appErrors.NewNotFoundError("Todo not found", map[string]interface{}{
				"todo_id": todoId,
			})
		}
		t.log.Error("Failed to find todo", err, map[string]interface{}{
			"todo_id": todoUUID,
		})
		return database.Todo{}, appErrors.NewInternalServerError(err, "Failed to find todo")
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
		t.log.Error("Failed to list todos", err, nil)
		return []database.Todo{}, appErrors.NewInternalServerError(err, "Failed to list todos")
	}

	// Log successful listing with count of todos
	t.log.Debug("Todos listed successfully", map[string]interface{}{
		"count": len(result),
	})
	return result, nil
}
