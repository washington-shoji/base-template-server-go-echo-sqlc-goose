package services

import (
	"go-echo-server-template/internal/database"
	appErrors "go-echo-server-template/internal/errors"
	"go-echo-server-template/internal/testutils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestService() (TodoService, *testutils.MockQueries) {
	mockDB := testutils.NewMockQueries()
	ctx := testutils.TestContext()
	return NewTodoService(ctx, mockDB), mockDB
}

func TestCreateTodo(t *testing.T) {
	service, _ := setupTestService()

	tests := []struct {
		name    string
		input   TodoParams
		wantErr bool
		errType error
	}{
		{
			name: "valid todo",
			input: TodoParams{
				Label:     "Test Todo",
				Completed: false,
			},
			wantErr: false,
		},
		{
			name: "empty label",
			input: TodoParams{
				Label:     "",
				Completed: false,
			},
			wantErr: true,
			errType: appErrors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateTodo(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					var appErr *appErrors.AppError
					assert.ErrorAs(t, err, &appErr)
					assert.Equal(t, tt.errType, appErr.Unwrap())
				}
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result.TodoID)
			assert.Equal(t, tt.input.Label, result.Label)
			assert.Equal(t, tt.input.Completed, result.Completed)
			assert.NotZero(t, result.CreatedAt)
		})
	}
}

func TestUpdateTodo(t *testing.T) {
	service, mockDB := setupTestService()

	// Create a test todo
	existingTodo := database.Todo{
		TodoID:    uuid.New(),
		Label:     "Existing Todo",
		Completed: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	mockDB.AddTestTodo(existingTodo)

	tests := []struct {
		name    string
		todoID  string
		input   TodoParams
		wantErr bool
		errType error
	}{
		{
			name:   "valid update",
			todoID: existingTodo.TodoID.String(),
			input: TodoParams{
				Label:     "Updated Todo",
				Completed: true,
			},
			wantErr: false,
		},
		{
			name:   "invalid uuid",
			todoID: "invalid-uuid",
			input: TodoParams{
				Label:     "Updated Todo",
				Completed: true,
			},
			wantErr: true,
			errType: appErrors.ErrBadRequest,
		},
		{
			name:   "todo not found",
			todoID: uuid.New().String(),
			input: TodoParams{
				Label:     "Updated Todo",
				Completed: true,
			},
			wantErr: true,
			errType: appErrors.ErrNotFound,
		},
		{
			name:   "empty label",
			todoID: existingTodo.TodoID.String(),
			input: TodoParams{
				Label:     "",
				Completed: true,
			},
			wantErr: true,
			errType: appErrors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add a small delay to ensure UpdatedAt will be different
			time.Sleep(1 * time.Millisecond)

			result, err := service.UpdateDoto(tt.todoID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					var appErr *appErrors.AppError
					assert.ErrorAs(t, err, &appErr)
					assert.Equal(t, tt.errType, appErr.Unwrap())
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, existingTodo.TodoID, result.TodoID)
			assert.Equal(t, tt.input.Label, result.Label)
			assert.Equal(t, tt.input.Completed, result.Completed)
			assert.True(t, result.UpdatedAt.After(existingTodo.UpdatedAt))
		})
	}
}

func TestDeleteTodo(t *testing.T) {
	service, mockDB := setupTestService()

	// Create a test todo
	existingTodo := database.Todo{
		TodoID:    uuid.New(),
		Label:     "Existing Todo",
		Completed: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	mockDB.AddTestTodo(existingTodo)

	tests := []struct {
		name    string
		todoID  string
		wantErr bool
		errType error
	}{
		{
			name:    "valid delete",
			todoID:  existingTodo.TodoID.String(),
			wantErr: false,
		},
		{
			name:    "invalid uuid",
			todoID:  "invalid-uuid",
			wantErr: true,
			errType: appErrors.ErrBadRequest,
		},
		{
			name:    "todo not found",
			todoID:  uuid.New().String(),
			wantErr: true,
			errType: appErrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteTodo(tt.todoID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					var appErr *appErrors.AppError
					assert.ErrorAs(t, err, &appErr)
					assert.Equal(t, tt.errType, appErr.Unwrap())
				}
				return
			}

			assert.NoError(t, err)
			// Verify todo was deleted
			_, err = mockDB.FindTodoById(testutils.TestContext(), existingTodo.TodoID)
			assert.Error(t, err)
		})
	}
}

func TestFindTodoById(t *testing.T) {
	service, mockDB := setupTestService()

	// Create a test todo
	existingTodo := database.Todo{
		TodoID:    uuid.New(),
		Label:     "Existing Todo",
		Completed: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	mockDB.AddTestTodo(existingTodo)

	tests := []struct {
		name    string
		todoID  string
		wantErr bool
		errType error
	}{
		{
			name:    "valid find",
			todoID:  existingTodo.TodoID.String(),
			wantErr: false,
		},
		{
			name:    "invalid uuid",
			todoID:  "invalid-uuid",
			wantErr: true,
			errType: appErrors.ErrBadRequest,
		},
		{
			name:    "todo not found",
			todoID:  uuid.New().String(),
			wantErr: true,
			errType: appErrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FindTodoById(tt.todoID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					var appErr *appErrors.AppError
					assert.ErrorAs(t, err, &appErr)
					assert.Equal(t, tt.errType, appErr.Unwrap())
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, existingTodo.TodoID, result.TodoID)
			assert.Equal(t, existingTodo.Label, result.Label)
			assert.Equal(t, existingTodo.Completed, result.Completed)
		})
	}
}

func TestListAllTodos(t *testing.T) {
	service, mockDB := setupTestService()

	// Create test todos
	todo1 := database.Todo{
		TodoID:    uuid.New(),
		Label:     "Todo 1",
		Completed: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	todo2 := database.Todo{
		TodoID:    uuid.New(),
		Label:     "Todo 2",
		Completed: true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	mockDB.AddTestTodo(todo1)
	mockDB.AddTestTodo(todo2)

	result, err := service.ListAllTodos()
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Clear todos and test empty list
	mockDB.ClearTodos()
	result, err = service.ListAllTodos()
	assert.NoError(t, err)
	assert.Empty(t, result)
}
