package todos_test

import (
	"context"
	"errors"
	"testing"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/todos"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = logging.Initialize("test", "ERROR")
}

func ctx() context.Context {
	return authn.WithPrincipal(context.Background(), authn.Principal{
		UserID: "u1", Roles: []string{"user"}, AuthMethod: "disabled",
	})
}

func TestCreateTodo(t *testing.T) {
	svc := todos.NewService(todos.NewFakeStore(), nil, true)
	todo, err := svc.CreateTodo(ctx(), todos.Params{Label: "hi"})
	require.NoError(t, err)
	assert.Equal(t, "hi", todo.Label)
	assert.False(t, todo.UpdatedAt.IsZero())

	_, err = svc.CreateTodo(ctx(), todos.Params{Label: ""})
	require.Error(t, err)
	assert.True(t, errors.Is(err, httpx.ErrValidation))
}

func TestUpdateTodo(t *testing.T) {
	store := todos.NewFakeStore()
	svc := todos.NewService(store, nil, true)
	id := uuid.New()
	store.Seed(todos.Todo{TodoID: id, Label: "old"})

	out, err := svc.UpdateTodo(ctx(), id.String(), todos.Params{Label: "new", Completed: true})
	require.NoError(t, err)
	assert.Equal(t, "new", out.Label)
	assert.True(t, out.Completed)

	_, err = svc.UpdateTodo(ctx(), "bad", todos.Params{Label: "x"})
	assert.True(t, errors.Is(err, httpx.ErrBadRequest))

	_, err = svc.UpdateTodo(ctx(), uuid.New().String(), todos.Params{Label: "x"})
	assert.True(t, errors.Is(err, httpx.ErrNotFound))
}

func TestDeleteAndGet(t *testing.T) {
	store := todos.NewFakeStore()
	svc := todos.NewService(store, nil, true)
	created, err := svc.CreateTodo(ctx(), todos.Params{Label: "x"})
	require.NoError(t, err)

	got, err := svc.FindTodoById(ctx(), created.TodoID.String())
	require.NoError(t, err)
	assert.Equal(t, created.TodoID, got.TodoID)

	require.NoError(t, svc.DeleteTodo(ctx(), created.TodoID.String()))
	_, err = svc.FindTodoById(ctx(), created.TodoID.String())
	assert.True(t, errors.Is(err, httpx.ErrNotFound))
}

func TestAuthRequired(t *testing.T) {
	svc := todos.NewService(todos.NewFakeStore(), nil, false)
	_, err := svc.CreateTodo(context.Background(), todos.Params{Label: "x"})
	assert.True(t, errors.Is(err, httpx.ErrPermissionDenied))
}
