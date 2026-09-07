package todos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/authz"
	"go-echo-server-template/internal/platform/events"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/logging"

	"github.com/google/uuid"
)

type Service struct {
	store        Store
	bus          *events.Bus
	authDisabled bool
}

func NewService(store Store, bus *events.Bus, authDisabled bool) *Service {
	return &Service{store: store, bus: bus, authDisabled: authDisabled}
}

func (s *Service) authorize(ctx context.Context) error {
	p, _ := authn.FromContext(ctx)
	if err := authz.RequireAuthenticated(p, s.authDisabled); err != nil {
		return fmt.Errorf("%w: authentication required", httpx.ErrPermissionDenied)
	}
	return nil
}

func (s *Service) CreateTodo(ctx context.Context, req Params) (Todo, error) {
	if err := s.authorize(ctx); err != nil {
		return Todo{}, err
	}
	log := logging.WithContext(ctx, "todos")
	if req.Label == "" {
		return Todo{}, fmt.Errorf("%w: label is required", httpx.ErrValidation)
	}
	now := time.Now().UTC()
	model := CreateParams{
		TodoID: uuid.New(), Label: req.Label, Completed: req.Completed,
		CreatedAt: now, UpdatedAt: now,
	}
	result, err := s.store.CreateTodo(ctx, model)
	if err != nil {
		log.Error("Failed to create todo", err, map[string]interface{}{"todo_id": model.TodoID})
		return Todo{}, fmt.Errorf("%w: failed to create todo", httpx.ErrInternal)
	}
	if s.bus != nil {
		s.bus.Publish(ctx, events.Event{Name: EventTodoCreated, Payload: result})
	}
	return result, nil
}

func (s *Service) UpdateTodo(ctx context.Context, todoID string, req Params) (Todo, error) {
	if err := s.authorize(ctx); err != nil {
		return Todo{}, err
	}
	if req.Label == "" {
		return Todo{}, fmt.Errorf("%w: label is required", httpx.ErrValidation)
	}
	id, err := uuid.Parse(todoID)
	if err != nil {
		return Todo{}, fmt.Errorf("%w: invalid todo id", httpx.ErrBadRequest)
	}
	result, err := s.store.UpdateTodo(ctx, UpdateParams{
		TodoID: id, Label: req.Label, Completed: req.Completed, UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Todo{}, fmt.Errorf("%w: todo not found", httpx.ErrNotFound)
		}
		return Todo{}, fmt.Errorf("%w: failed to update todo", httpx.ErrInternal)
	}
	return result, nil
}

func (s *Service) DeleteTodo(ctx context.Context, todoID string) error {
	if err := s.authorize(ctx); err != nil {
		return err
	}
	id, err := uuid.Parse(todoID)
	if err != nil {
		return fmt.Errorf("%w: invalid todo id", httpx.ErrBadRequest)
	}
	// Ensure exists
	if _, err := s.store.FindTodoById(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: todo not found", httpx.ErrNotFound)
		}
		return fmt.Errorf("%w: failed to delete todo", httpx.ErrInternal)
	}
	if err := s.store.DeleteTodo(ctx, id); err != nil {
		return fmt.Errorf("%w: failed to delete todo", httpx.ErrInternal)
	}
	return nil
}

func (s *Service) FindTodoById(ctx context.Context, todoID string) (Todo, error) {
	if err := s.authorize(ctx); err != nil {
		return Todo{}, err
	}
	id, err := uuid.Parse(todoID)
	if err != nil {
		return Todo{}, fmt.Errorf("%w: invalid todo id", httpx.ErrBadRequest)
	}
	result, err := s.store.FindTodoById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Todo{}, fmt.Errorf("%w: todo not found", httpx.ErrNotFound)
		}
		return Todo{}, fmt.Errorf("%w: failed to find todo", httpx.ErrInternal)
	}
	return result, nil
}

func (s *Service) ListAllTodos(ctx context.Context) ([]Todo, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	result, err := s.store.ListAllTodos(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list todos", httpx.ErrInternal)
	}
	return result, nil
}

// Summary returns global todo aggregates for cross-domain consumers.
func (s *Service) Summary(ctx context.Context) (Summary, error) {
	if err := s.authorize(ctx); err != nil {
		return Summary{}, err
	}
	result, err := s.store.CountTodos(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("%w: failed to summarise todos", httpx.ErrInternal)
	}
	return result, nil
}
