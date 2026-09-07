package api

import (
	"net/http"

	"go-echo-server-template/internal/auth"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/httpx"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *auth.Service
	cfg config.AuthConfig
}

func Register(e *echo.Echo, svc *auth.Service, cfg config.AuthConfig) {
	h := &Handler{svc: svc, cfg: cfg}
	api := e.Group("/api/v1/auth")
	api.POST("/register", h.Register)
	api.POST("/login", h.Login)
	api.POST("/logout", h.Logout)
	api.POST("/token", h.CreateToken)
}

type creds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(c echo.Context) error {
	var req creds
	if err := c.Bind(&req); err != nil {
		return httpx.NewAPIError(httpx.ErrBadRequest, "invalid body", nil)
	}
	u, err := h.svc.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]string{"user_id": u.UserID.String(), "email": u.Email})
}

func (h *Handler) Login(c echo.Context) error {
	var req creds
	if err := c.Bind(&req); err != nil {
		return httpx.NewAPIError(httpx.ErrBadRequest, "invalid body", nil)
	}
	sess, u, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	c.SetCookie(&http.Cookie{
		Name:     h.cfg.SessionCookie,
		Value:    sess.SessionID.String(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  sess.ExpiresAt,
	})
	return c.JSON(http.StatusOK, map[string]string{"user_id": u.UserID.String(), "email": u.Email})
}

func (h *Handler) Logout(c echo.Context) error {
	if cookie, err := c.Cookie(h.cfg.SessionCookie); err == nil {
		if id, err := uuid.Parse(cookie.Value); err == nil {
			_ = h.svc.Logout(c.Request().Context(), id)
		}
	}
	c.SetCookie(&http.Cookie{Name: h.cfg.SessionCookie, Value: "", Path: "/", MaxAge: -1})
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) CreateToken(c echo.Context) error {
	var req creds
	if err := c.Bind(&req); err != nil {
		return httpx.NewAPIError(httpx.ErrBadRequest, "invalid body", nil)
	}
	_, u, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	token, err := h.svc.CreateBearerToken(c.Request().Context(), u.UserID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"token": token, "token_type": "Bearer"})
}
