package services

import (
	"go-echo-server-template/internal/database"
)

type TodoParams struct {
	Label     string `json:"label"`
	Completed bool   `json:"completed"`
}

type TodoService interface {
	CreateTodo(reqModel TodoParams) (database.Todo, error)
	UpdateDoto(todoId string, reqModel TodoParams) (database.Todo, error)
	DeleteTodo(todoId string) error
	FindTodoById(todoId string) (database.Todo, error)
	ListAllTodos() ([]database.Todo, error)
}
