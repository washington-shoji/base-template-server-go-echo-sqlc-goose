package api

import (
	"net/http"

	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/todos"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *todos.Service
}

func NewHandler(svc *todos.Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts legacy and resource-oriented Todo routes.
func Register(e *echo.Echo, svc *todos.Service) {
	h := NewHandler(svc)
	api := e.Group("/api/v1")

	// Canonical resource routes
	api.POST("/todos", h.Create)
	api.GET("/todos", h.List)
	api.GET("/todos/:id", h.Get)
	api.PUT("/todos/:id", h.Update)
	api.DELETE("/todos/:id", h.Delete)

	// Legacy aliases (preserved for compatibility)
	api.POST("/create-todo", h.Create)
	api.PUT("/update-todo/:todo-id", h.UpdateLegacy)
	api.DELETE("/delete-todo/:todo-id", h.DeleteLegacy)
	api.GET("/todo/:todo-id", h.GetLegacy)
}

func (h *Handler) Create(c echo.Context) error {
	var req todos.Params
	if err := c.Bind(&req); err != nil {
		return httpx.NewAPIError(httpx.ErrBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
	}
	out, err := h.svc.CreateTodo(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) List(c echo.Context) error {
	out, err := h.svc.ListAllTodos(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) Get(c echo.Context) error {
	out, err := h.svc.FindTodoById(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) Update(c echo.Context) error {
	var req todos.Params
	if err := c.Bind(&req); err != nil {
		return httpx.NewAPIError(httpx.ErrBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
	}
	out, err := h.svc.UpdateTodo(c.Request().Context(), c.Param("id"), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.svc.DeleteTodo(c.Request().Context(), c.Param("id")); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "todo deleted successfully"})
}

func (h *Handler) GetLegacy(c echo.Context) error {
	out, err := h.svc.FindTodoById(c.Request().Context(), c.Param("todo-id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) UpdateLegacy(c echo.Context) error {
	var req todos.Params
	if err := c.Bind(&req); err != nil {
		return httpx.NewAPIError(httpx.ErrBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
	}
	out, err := h.svc.UpdateTodo(c.Request().Context(), c.Param("todo-id"), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteLegacy(c echo.Context) error {
	if err := h.svc.DeleteTodo(c.Request().Context(), c.Param("todo-id")); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "todo deleted successfully"})
}
