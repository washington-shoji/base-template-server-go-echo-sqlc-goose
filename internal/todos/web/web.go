package web

import (
	"errors"
	"net/http"
	"strings"

	"go-echo-server-template/internal/auth"
	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/todos"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *todos.Service
	cfg config.AuthConfig
}

func Register(e *echo.Echo, svc *todos.Service, cfg config.AuthConfig) {
	h := &Handler{svc: svc, cfg: cfg}
	app := e.Group("/app")
	app.GET("/todos", h.List)
	app.POST("/todos", h.Create)
	app.DELETE("/todos/:id", h.Delete)
}

type todoForm struct {
	Label string
}

type listPage struct {
	httpx.Shell
	Todos  []todos.Todo
	Form   todoForm
	Errors httpx.FormErrors
}

func (h *Handler) csrf(c echo.Context) string {
	if cookie, err := c.Cookie(h.cfg.CSRFCookieName); err == nil {
		return cookie.Value
	}
	token := auth.NewCSRFToken()
	c.SetCookie(&http.Cookie{Name: h.cfg.CSRFCookieName, Value: token, Path: "/", SameSite: http.SameSiteLaxMode})
	return token
}

func (h *Handler) user(c echo.Context) string {
	if p, ok := authn.FromContext(c.Request().Context()); ok && p.UserID != "" {
		return p.UserID
	}
	return ""
}

func (h *Handler) page(c echo.Context, items []todos.Todo, form todoForm, errs httpx.FormErrors) listPage {
	return listPage{
		Shell: httpx.Shell{
			Title: "Todos",
			CSRF:  h.csrf(c),
			Flash: httpx.ConsumeFlash(c),
			User:  h.user(c),
			Nav:   "todos",
		},
		Todos:  items,
		Form:   form,
		Errors: errs,
	}
}

func (h *Handler) List(c echo.Context) error {
	items, err := h.svc.ListAllTodos(c.Request().Context())
	if err != nil {
		return err
	}
	data := h.page(c, items, todoForm{}, httpx.FormErrors{Fields: map[string]string{}})
	if httpx.IsHTMX(c) {
		return httpx.RenderFragment(c, "fragments/todos/list.html", data)
	}
	return httpx.RenderApp(c, "pages/todos/list.html", data)
}

func (h *Handler) Create(c echo.Context) error {
	label := strings.TrimSpace(c.FormValue("label"))
	form := todoForm{Label: label}
	_, err := h.svc.CreateTodo(c.Request().Context(), todos.Params{Label: label})
	if err != nil {
		errs := httpx.FormErrors{Fields: map[string]string{}}
		if errors.Is(err, httpx.ErrValidation) {
			errs.Fields["label"] = "Label is required"
			if msg := err.Error(); strings.Contains(msg, ":") {
				errs.Fields["label"] = strings.TrimSpace(msg[strings.LastIndex(msg, ":")+1:])
			}
		} else {
			errs.Global = "Could not create todo"
		}
		items, _ := h.svc.ListAllTodos(c.Request().Context())
		data := h.page(c, items, form, errs)
		if httpx.IsHTMX(c) {
			return httpx.RenderFragmentStatus(c, http.StatusUnprocessableEntity, "pages/todos/list.html", data)
		}
		return httpx.RenderAppStatus(c, http.StatusUnprocessableEntity, "pages/todos/list.html", data)
	}
	if httpx.IsHTMX(c) {
		items, err := h.svc.ListAllTodos(c.Request().Context())
		if err != nil {
			return err
		}
		data := h.page(c, items, todoForm{}, httpx.FormErrors{Fields: map[string]string{}})
		return httpx.RenderFragment(c, "fragments/todos/list.html", data)
	}
	httpx.FlashSuccess(c, "Todo created")
	return c.Redirect(http.StatusSeeOther, "/app/todos")
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.svc.DeleteTodo(c.Request().Context(), c.Param("id")); err != nil {
		return err
	}
	httpx.FlashSuccess(c, "Todo deleted")
	return h.List(c)
}
