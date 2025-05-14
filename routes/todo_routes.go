package routes

import (
	"context"
	"go-echo-server-template/handlers"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/services"

	"github.com/labstack/echo/v4"
)

// RegisterTodoRoutes registers all todo-related routes
func RegisterTodoRoutes(e *echo.Echo, queries *database.Queries) {
	// Create service and handler
	ctx := context.Background()
	todoService := services.NewTodoService(ctx, queries)
	todoHandler := handlers.NewTodoHandler(ctx, todoService)

	// API group
	api := e.Group("/api/v1")

	// Todo routes
	api.POST("/create-todo", todoHandler.CreateTodoHandler)
	api.PUT("/update-todo/:todo-id", todoHandler.UpdateTodoHandler)
	api.DELETE("/delete-todo/:todo-id", todoHandler.DeleteTodoHandler)
	api.GET("/todo/:todo-id", todoHandler.FindTodoByIdHandler)
	api.GET("/todos", todoHandler.ListAllTodosHandler)
}
