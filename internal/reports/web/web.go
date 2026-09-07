package web

import (
	"net/http"

	"go-echo-server-template/internal/auth"
	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/reports"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *reports.Service
	cfg config.AuthConfig
}

func Register(e *echo.Echo, svc *reports.Service, cfg config.AuthConfig) {
	h := &Handler{svc: svc, cfg: cfg}
	app := e.Group("/app")
	app.GET("/reports", h.List)
	app.POST("/reports", h.Create)
	app.GET("/reports/:id", h.Show)
	app.GET("/reports/:id/status", h.StatusFragment)
}

type listPage struct {
	httpx.Shell
	Reports []reports.Report
}

type showPage struct {
	httpx.Shell
	Report reports.Report
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

func (h *Handler) List(c echo.Context) error {
	items, err := h.svc.ListReports(c.Request().Context())
	if err != nil {
		return err
	}
	data := listPage{
		Shell: httpx.Shell{
			Title: "Reports",
			CSRF:  h.csrf(c),
			Flash: httpx.ConsumeFlash(c),
			User:  h.user(c),
			Nav:   "reports",
		},
		Reports: items,
	}
	if httpx.IsHTMX(c) {
		return httpx.RenderFragment(c, "fragments/reports/list.html", data)
	}
	return httpx.RenderApp(c, "pages/reports/list.html", data)
}

func (h *Handler) Create(c echo.Context) error {
	r, err := h.svc.CreateReport(c.Request().Context(), reports.CreateRequest{
		Type: c.FormValue("type"),
	})
	if err != nil {
		return err
	}
	httpx.FlashInfo(c, "Report generation started")
	if httpx.IsHTMX(c) {
		return h.List(c)
	}
	return c.Redirect(http.StatusSeeOther, "/app/reports/"+r.ReportID.String())
}

func (h *Handler) Show(c echo.Context) error {
	r, err := h.svc.GetReport(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	data := showPage{
		Shell: httpx.Shell{
			Title: "Report",
			CSRF:  h.csrf(c),
			Flash: httpx.ConsumeFlash(c),
			User:  h.user(c),
			Nav:   "reports",
		},
		Report: r,
	}
	return httpx.RenderApp(c, "pages/reports/show.html", data)
}

func (h *Handler) StatusFragment(c echo.Context) error {
	r, err := h.svc.GetReport(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return httpx.RenderFragment(c, "fragments/reports/status.html", showPage{Report: r})
}
