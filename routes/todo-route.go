package routes

import (
	"context"
	"go-echo-server-template/handlers"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/services"

	"github.com/labstack/echo/v4"
)

func InitTodoRouter(e *echo.Echo, ctx context.Context, db *database.Queries) {
	todoService := services.NewTodoService(ctx, db)
	todoHandler := handlers.NewTodoHandler(ctx, todoService)

	group := e.Group("api/v1")

	group.POST("/create-todo", todoHandler.CreateTodoHandler)
	group.PUT("/update-todo/:todo-id", todoHandler.UpdateTodoHandler)
	group.DELETE("/delete-todo/:todo-id", todoHandler.DeleteTodoHandler)
	group.GET("/todo/:todo-id", todoHandler.FindTodoByIdHandler)
	group.GET("/todos", todoHandler.ListAllTodosHandler)
}
