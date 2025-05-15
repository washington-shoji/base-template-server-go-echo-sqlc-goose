package handlers

import (
	"context"
	appErrors "go-echo-server-template/internal/errors"
	"go-echo-server-template/internal/logger"
	"go-echo-server-template/services"
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

	todoModel := services.TodoParams{}
	if err := ctx.Bind(&todoModel); err != nil {
		return appErrors.NewBadRequestError("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
	}

	response, err := h.TodoService.CreateTodo(todoModel)
	if err != nil {
		return err // Our error handler will handle the service errors
	}

	return ctx.JSON(http.StatusCreated, response)
}

func (h *TodoHandler) UpdateTodoHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return appErrors.NewBadRequestError("Todo ID is required", nil)
	}

	todoModel := services.TodoParams{}
	if err := ctx.Bind(&todoModel); err != nil {
		return appErrors.NewBadRequestError("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
	}

	response, err := h.TodoService.UpdateDoto(todoId, todoModel)
	if err != nil {
		return err // Our error handler will handle the service errors
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *TodoHandler) DeleteTodoHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return appErrors.NewBadRequestError("Todo ID is required", nil)
	}

	err := h.TodoService.DeleteTodo(todoId)
	if err != nil {
		return err // Our error handler will handle the service errors
	}

	return ctx.JSON(http.StatusOK, successResponse{Message: "todo deleted successfully"})
}

func (h *TodoHandler) FindTodoByIdHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return appErrors.NewBadRequestError("Todo ID is required", nil)
	}

	response, err := h.TodoService.FindTodoById(todoId)
	if err != nil {
		return err // Our error handler will handle the service errors
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *TodoHandler) ListAllTodosHandler(ctx echo.Context) error {
	// Initialize request-specific logger
	h.log = logger.WithContext(ctx.Request().Context(), "todo_handler")

	response, err := h.TodoService.ListAllTodos()
	if err != nil {
		return err // Our error handler will handle the service errors
	}

	return ctx.JSON(http.StatusOK, response)
}
