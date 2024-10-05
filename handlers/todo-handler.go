package handlers

import (
	"go-echo-server-template/services"
	"go-echo-server-template/utils"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type successResponse struct {
	Message string `json:"message"`
}

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
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	response, err := h.TodoService.CreateTodo(todoModel)
	if err != nil {
		log.Printf("service error: %v", err)
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not create todo")

	}

	return utils.RespondWithJSON(ctx, http.StatusCreated, response)
}

func (h *TodoHandler) UpdateTodoHandler(ctx echo.Context) error {
	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	todoModel := services.TodoParams{}
	err := ctx.Bind(&todoModel)
	if err != nil {
		log.Printf("binding error: %v", err)
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	response, err := h.TodoService.UpdateDoto(todoId, todoModel)
	if err != nil {
		log.Printf("service error: %v", err)
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not update todo")
	}

	return utils.RespondWithJSON(ctx, http.StatusOK, response)
}

func (h *TodoHandler) DeleteTodoHandler(ctx echo.Context) error {
	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	err := h.TodoService.DeleteTodo(todoId)
	if err != nil {
		log.Printf("service error: %v", err)
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not delete todo")
	}

	return utils.RespondWithJSON(ctx, http.StatusOK, successResponse{Message: "todo deleted successfully"})

}

func (h *TodoHandler) FindTodoByIdHandler(ctx echo.Context) error {
	todoId := ctx.Param("todo-id")
	if todoId == "" {
		return utils.RespondWithError(ctx, http.StatusBadRequest, "bad request data malformed")
	}

	response, err := h.TodoService.FindTodoById(todoId)
	if err != nil {
		log.Printf("service error: %v", err)
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not find todo")
	}

	return utils.RespondWithJSON(ctx, http.StatusOK, response)
}

func (h *TodoHandler) ListAllTodosHandler(ctx echo.Context) error {

	response, err := h.TodoService.ListAllTodos()
	if err != nil {
		log.Printf("service error: %v", err)
		return utils.RespondWithError(ctx, http.StatusInternalServerError, "could not fetch todos")
	}

	return utils.RespondWithJSON(ctx, http.StatusOK, response)
}
