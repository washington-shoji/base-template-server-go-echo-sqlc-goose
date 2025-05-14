package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-echo-server-template/internal/database"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type todoRequest struct {
	Label     string `json:"label"`
	Completed bool   `json:"completed"`
}

func TestTodoAPIIntegration(t *testing.T) {
	// Test Create Todo
	t.Run("Create Todo", func(t *testing.T) {
		reqBody := todoRequest{
			Label:     "Test Todo",
			Completed: false,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/create-todo", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		testServer.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response database.Todo
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, reqBody.Label, response.Label)
		assert.Equal(t, reqBody.Completed, response.Completed)
		assert.NotEmpty(t, response.TodoID)

		// Test Get Todo
		t.Run("Get Created Todo", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/todo/%s", response.TodoID), nil)
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var getTodoResponse database.Todo
			err := json.Unmarshal(rec.Body.Bytes(), &getTodoResponse)
			assert.NoError(t, err)
			assert.Equal(t, response.TodoID, getTodoResponse.TodoID)
			assert.Equal(t, reqBody.Label, getTodoResponse.Label)
		})

		// Test Update Todo
		t.Run("Update Todo", func(t *testing.T) {
			updateReq := todoRequest{
				Label:     "Updated Todo",
				Completed: true,
			}
			jsonBody, _ := json.Marshal(updateReq)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/update-todo/%s", response.TodoID), bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var updateResponse database.Todo
			err := json.Unmarshal(rec.Body.Bytes(), &updateResponse)
			assert.NoError(t, err)
			assert.Equal(t, response.TodoID, updateResponse.TodoID)
			assert.Equal(t, updateReq.Label, updateResponse.Label)
			assert.Equal(t, updateReq.Completed, updateResponse.Completed)
			assert.True(t, updateResponse.UpdatedAt.After(response.UpdatedAt))
		})

		// Test List Todos
		t.Run("List Todos", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var todos []database.Todo
			err := json.Unmarshal(rec.Body.Bytes(), &todos)
			assert.NoError(t, err)
			assert.NotEmpty(t, todos)
		})

		// Test Delete Todo
		t.Run("Delete Todo", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/delete-todo/%s", response.TodoID), nil)
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify todo is deleted
			req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/todo/%s", response.TodoID), nil)
			rec = httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	})

	// Test Error Cases
	t.Run("Error Cases", func(t *testing.T) {
		// Test Invalid UUID
		t.Run("Get Todo with Invalid UUID", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/todo/invalid-uuid", nil)
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		// Test Not Found
		t.Run("Get Non-existent Todo", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/todo/123e4567-e89b-12d3-a456-426614174000", nil)
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		// Test Invalid Input
		t.Run("Create Todo with Empty Label", func(t *testing.T) {
			reqBody := todoRequest{
				Label:     "",
				Completed: false,
			}
			jsonBody, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/create-todo", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		})
	})
}
