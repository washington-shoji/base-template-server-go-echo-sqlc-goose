package web

import (
	"net/http"

	"go-echo-server-template/internal/auth"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/httpx"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *auth.Service
	cfg config.AuthConfig
}

func Register(e *echo.Echo, svc *auth.Service, cfg config.AuthConfig) {
	h := &Handler{svc: svc, cfg: cfg}
	e.GET("/app/login", h.LoginForm)
	e.POST("/app/login", h.Login)
}

type loginPage struct {
	httpx.Shell
	Error string
	Email string
}

func (h *Handler) LoginForm(c echo.Context) error {
	csrf := auth.NewCSRFToken()
	c.SetCookie(&http.Cookie{Name: h.cfg.CSRFCookieName, Value: csrf, Path: "/", SameSite: http.SameSiteLaxMode})
	return httpx.RenderAuth(c, "pages/auth/login.html", loginPage{
		Shell: httpx.Shell{Title: "Sign in", CSRF: csrf, Flash: httpx.ConsumeFlash(c), Nav: "login"},
	})
}

func (h *Handler) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	sess, _, err := h.svc.Login(c.Request().Context(), email, password)
	if err != nil {
		csrf := auth.NewCSRFToken()
		c.SetCookie(&http.Cookie{Name: h.cfg.CSRFCookieName, Value: csrf, Path: "/", SameSite: http.SameSiteLaxMode})
		return httpx.RenderAuthStatus(c, http.StatusUnauthorized, "pages/auth/login.html", loginPage{
			Shell: httpx.Shell{Title: "Sign in", CSRF: csrf, Nav: "login"},
			Error: "Invalid credentials",
			Email: email,
		})
	}
	c.SetCookie(&http.Cookie{
		Name: h.cfg.SessionCookie, Value: sess.SessionID.String(), Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Expires: sess.ExpiresAt,
	})
	httpx.FlashSuccess(c, "Signed in")
	return c.Redirect(http.StatusSeeOther, "/app/todos")
}
