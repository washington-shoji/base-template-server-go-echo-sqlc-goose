package handlers

import (
	"go-echo-server-template/services"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	TodoService services.TodoService
}

func NewTodoHandler(todoService services.TodoService) *TodoHandler {
	return &TodoHandler{
		TodoService: todoService,
	}
}

func (h *TodoHandler) CreateTodoHandler(ctx echo.Context) error {
	todoModel := services.TodoParams{}
	err := ctx.Bind(&todoModel)
	if err != nil {
		log.Printf("binding error: %v", err)
		return ctx.String(http.StatusBadRequest, "bad request - data malformed")
	}

	response, err := h.TodoService.CreateTodo(todoModel)
	if err != nil {
		log.Printf("service error: %v", err)
		return ctx.String(http.StatusInternalServerError, "could not create todo")
	}

	return ctx.JSON(http.StatusCreated, response)
}

func (h *TodoHandler) UpdateTodoHandler(ctx echo.Context) error {
	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return ctx.String(http.StatusBadRequest, "bad request - data malformed")
	}

	todoModel := services.TodoParams{}
	err := ctx.Bind(&todoModel)
	if err != nil {
		log.Printf("binding error: %v", err)
		return ctx.String(http.StatusBadRequest, "bad request - data malformed")
	}

	response, err := h.TodoService.UpdateDoto(todoId, todoModel)
	if err != nil {
		log.Printf("service error: %v", err)
		return ctx.String(http.StatusInternalServerError, "could not create todo")
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *TodoHandler) DeleteTodoHandler(ctx echo.Context) error {
	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return ctx.String(http.StatusBadRequest, "bad request - data malformed")
	}

	err := h.TodoService.DeleteTodo(todoId)
	if err != nil {
		log.Printf("service error: %v", err)
		return ctx.String(http.StatusInternalServerError, "could not delte todo")
	}

	return ctx.String(http.StatusOK, "todo deleted successfully")
}

func (h *TodoHandler) FindTodoByIdHandler(ctx echo.Context) error {
	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return ctx.String(http.StatusBadRequest, "bad request - data malformed")
	}

	response, err := h.TodoService.FindTodoById(todoId)
	if err != nil {
		log.Printf("service error: %v", err)
		return ctx.String(http.StatusInternalServerError, "could not find todo")
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *TodoHandler) ListAllTodosHandler(ctx echo.Context) error {

	response, err := h.TodoService.ListAllTodos()
	if err != nil {
		log.Printf("service error: %v", err)
		return ctx.String(http.StatusInternalServerError, "could not fetch todos")
	}

	return ctx.JSON(http.StatusOK, response)
}
