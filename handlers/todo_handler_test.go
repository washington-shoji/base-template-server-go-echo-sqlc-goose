package handlers

import (
	"encoding/json"
	"go-echo-server-template/internal/database"
	appErrors "go-echo-server-template/internal/errors"
	"go-echo-server-template/internal/testutils"
	"go-echo-server-template/services"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() (*TodoHandler, *testutils.MockQueries) {
	mockDB := testutils.NewMockQueries()
	ctx := testutils.TestContext()
	todoService := services.NewTodoService(ctx, mockDB)
	return NewTodoHandler(ctx, todoService), mockDB
}

func TestCreateTodoHandler(t *testing.T) {
	handler, _ := setupTestHandler()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "valid todo",
			body:       `{"label": "Test Todo", "completed": false}`,
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "invalid json",
			body:       `{"label": "Test Todo", "completed": }`,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "empty label",
			body:       `{"label": "", "completed": false}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := testutils.MockContext(http.MethodPost, "/api/todos", tt.body)
			err := handler.CreateTodoHandler(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *appErrors.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, tt.wantStatus, appErr.StatusCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var response database.Todo
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.TodoID)
			assert.NotEmpty(t, response.Label)
		})
	}
}

func TestUpdateTodoHandler(t *testing.T) {
	handler, mockDB := setupTestHandler()

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
		name       string
		todoID     string
		body       string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "valid update",
			todoID:     existingTodo.TodoID.String(),
			body:       `{"label": "Updated Todo", "completed": true}`,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "invalid uuid",
			todoID:     "invalid-uuid",
			body:       `{"label": "Updated Todo", "completed": true}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "todo not found",
			todoID:     uuid.New().String(),
			body:       `{"label": "Updated Todo", "completed": true}`,
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "invalid json",
			todoID:     existingTodo.TodoID.String(),
			body:       `{"label": "Updated Todo", "completed": }`,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := testutils.MockContextWithParams(
				http.MethodPut,
				"/api/todos/:todo-id",
				tt.body,
				map[string]string{"todo-id": tt.todoID},
			)
			err := handler.UpdateTodoHandler(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *appErrors.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, tt.wantStatus, appErr.StatusCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var response database.Todo
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, existingTodo.TodoID, response.TodoID)
		})
	}
}

func TestDeleteTodoHandler(t *testing.T) {
	handler, mockDB := setupTestHandler()

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
		name       string
		todoID     string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "valid delete",
			todoID:     existingTodo.TodoID.String(),
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "invalid uuid",
			todoID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "todo not found",
			todoID:     uuid.New().String(),
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := testutils.MockContextWithParams(
				http.MethodDelete,
				"/api/todos/:todo-id",
				"",
				map[string]string{"todo-id": tt.todoID},
			)
			err := handler.DeleteTodoHandler(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *appErrors.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, tt.wantStatus, appErr.StatusCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var response successResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "todo deleted successfully", response.Message)
		})
	}
}

func TestFindTodoByIdHandler(t *testing.T) {
	handler, mockDB := setupTestHandler()

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
		name       string
		todoID     string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "valid find",
			todoID:     existingTodo.TodoID.String(),
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "invalid uuid",
			todoID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "todo not found",
			todoID:     uuid.New().String(),
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := testutils.MockContextWithParams(
				http.MethodGet,
				"/api/todos/:todo-id",
				"",
				map[string]string{"todo-id": tt.todoID},
			)
			err := handler.FindTodoByIdHandler(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *appErrors.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, tt.wantStatus, appErr.StatusCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var response database.Todo
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, existingTodo.TodoID, response.TodoID)
		})
	}
}

func TestListAllTodosHandler(t *testing.T) {
	handler, mockDB := setupTestHandler()

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

	ctx, rec := testutils.MockContext(http.MethodGet, "/api/todos", "")
	err := handler.ListAllTodosHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response []database.Todo
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	// Test empty list
	mockDB.ClearTodos()
	ctx, rec = testutils.MockContext(http.MethodGet, "/api/todos", "")
	err = handler.ListAllTodosHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response)
}
