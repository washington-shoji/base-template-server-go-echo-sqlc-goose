package handlers

import (
	"context"
	"go-echo-server-template/internal/logger"
	"go-echo-server-template/services"
	"go-echo-server-template/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type successResponse struct {
	Message string `json:"message"`
}

type TodoHandler struct {
	TodoService services.TodoService
	log         *logger.Logger
}

func NewTodoHandler(ctx context.Context, todoService services.TodoService) *TodoHandler {
	return &TodoHandler{
		TodoService: todoService,
		log:         logger.WithContext(ctx, "todo_handler"),
	}
}

func (h *TodoHandler) CreateTodoHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	// Log the start of todo creation request
	h.log.Debug("Received create todo request", map[string]interface{}{
		"method": ctx.Request().Method,
		"path":   ctx.Request().URL.Path,
	})

	todoModel := services.TodoParams{}
	err := ctx.Bind(&todoModel)
	if err != nil {
		// Log request binding error
		h.log.Error("Failed to bind request body", err, map[string]interface{}{
			"error_type": "binding_error",
		})
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	response, err := h.TodoService.CreateTodo(todoModel)
	if err != nil {
		// Log service layer error
		h.log.Error("Service failed to create todo", err, map[string]interface{}{
			"error_type": "service_error",
			"label":      todoModel.Label,
		})
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not create todo")
	}

	// Log successful todo creation
	h.log.Debug("Todo created successfully", map[string]interface{}{
		"todo_id": response.TodoID,
		"status":  http.StatusCreated,
	})
	return utils.RespondWithJSON(ctx, http.StatusCreated, response)
}

func (h *TodoHandler) UpdateTodoHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	// Log the start of todo update request
	h.log.Debug("Received update todo request", map[string]interface{}{
		"method": ctx.Request().Method,
		"path":   ctx.Request().URL.Path,
	})

	todoId := ctx.Param("todo-id")
	if todoId == "" {
		// Log missing todo ID error
		h.log.Error("Missing todo ID in request", nil, map[string]interface{}{
			"error_type": "validation_error",
		})
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	todoModel := services.TodoParams{}
	err := ctx.Bind(&todoModel)
	if err != nil {
		// Log request binding error
		h.log.Error("Failed to bind request body", err, map[string]interface{}{
			"error_type": "binding_error",
			"todo_id":    todoId,
		})
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	response, err := h.TodoService.UpdateDoto(todoId, todoModel)
	if err != nil {
		// Log service layer error
		h.log.Error("Service failed to update todo", err, map[string]interface{}{
			"error_type": "service_error",
			"todo_id":    todoId,
		})
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not update todo")
	}

	// Log successful todo update
	h.log.Debug("Todo updated successfully", map[string]interface{}{
		"todo_id": response.TodoID,
		"status":  http.StatusOK,
	})
	return utils.RespondWithJSON(ctx, http.StatusOK, response)
}

func (h *TodoHandler) DeleteTodoHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	// Log the start of todo deletion request
	h.log.Debug("Received delete todo request", map[string]interface{}{
		"method": ctx.Request().Method,
		"path":   ctx.Request().URL.Path,
	})

	todoId := ctx.Param("todo-id")
	if todoId == "" {
		// Log missing todo ID error
		h.log.Error("Missing todo ID in request", nil, map[string]interface{}{
			"error_type": "validation_error",
		})
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	err := h.TodoService.DeleteTodo(todoId)
	if err != nil {
		// Log service layer error
		h.log.Error("Service failed to delete todo", err, map[string]interface{}{
			"error_type": "service_error",
			"todo_id":    todoId,
		})
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not delete todo")
	}

	// Log successful todo deletion
	h.log.Debug("Todo deleted successfully", map[string]interface{}{
		"todo_id": todoId,
		"status":  http.StatusOK,
	})
	return utils.RespondWithJSON(ctx, http.StatusOK, successResponse{Message: "todo deleted successfully"})
}

func (h *TodoHandler) FindTodoByIdHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	// Log the start of todo search request
	h.log.Debug("Received find todo request", map[string]interface{}{
		"method": ctx.Request().Method,
		"path":   ctx.Request().URL.Path,
	})

	todoId := ctx.Param("todo-id")
	if todoId == "" {
		// Log missing todo ID error
		h.log.Error("Missing todo ID in request", nil, map[string]interface{}{
			"error_type": "validation_error",
		})
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	response, err := h.TodoService.FindTodoById(todoId)
	if err != nil {
		// Log service layer error
		h.log.Error("Service failed to find todo", err, map[string]interface{}{
			"error_type": "service_error",
			"todo_id":    todoId,
		})
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not find todo")
	}

	// Log successful todo retrieval
	h.log.Debug("Todo found successfully", map[string]interface{}{
		"todo_id": response.TodoID,
		"status":  http.StatusOK,
	})
	return utils.RespondWithJSON(ctx, http.StatusOK, response)
}

func (h *TodoHandler) ListAllTodosHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	// Log the start of list todos request
	h.log.Debug("Received list todos request", map[string]interface{}{
		"method": ctx.Request().Method,
		"path":   ctx.Request().URL.Path,
	})

	response, err := h.TodoService.ListAllTodos()
	if err != nil {
		// Log service layer error
		h.log.Error("Service failed to list todos", err, map[string]interface{}{
			"error_type": "service_error",
		})
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not fetch todos")
	}

	// Log successful todos listing
	h.log.Debug("Todos listed successfully", map[string]interface{}{
		"count":  len(response),
		"status": http.StatusOK,
	})
	return utils.RespondWithJSON(ctx, http.StatusOK, response)
}
