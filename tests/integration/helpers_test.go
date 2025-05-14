package integration

import (
	"context"
	"go-echo-server-template/internal/database"
	"time"

	"github.com/google/uuid"
)

// createTestTodo creates a todo in the database for testing
func createTestTodo(label string, completed bool) (database.Todo, error) {
	todo := database.CreateTodoParams{
		TodoID:    uuid.New(),
		Label:     label,
		Completed: completed,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return queries.CreateTodo(context.Background(), todo)
}

// cleanupTestData removes all test data from the database
func cleanupTestData() error {
	todos, err := queries.ListAllTodos(context.Background())
	if err != nil {
		return err
	}

	for _, todo := range todos {
		if err := queries.DeleteTodo(context.Background(), todo.TodoID); err != nil {
			return err
		}
	}

	return nil
}
