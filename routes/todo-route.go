package routes

import (
	"context"
	"go-echo-server-template/handlers"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/services"

	"github.com/labstack/echo/v4"
)

func InitTodoRouter(e *echo.Echo, ctx context.Context, q *database.Queries) {
	todoService := services.NewTodoService(ctx, q)
	handler := handlers.NewTodoHandler(todoService)

	group := e.Group("api/v1")

	group.POST("/create-todo", handler.CreateTodoHandler)
	group.PUT("/update-todo/:todo-id", handler.UpdateTodoHandler)
	group.DELETE("delete-todo/:todo-id", handler.DeleteTodoHandler)
	group.GET("/todo/:todo-id", handler.FindTodoByIdHandler)
	group.GET("/todo", handler.ListAllTodosHandler)
}
