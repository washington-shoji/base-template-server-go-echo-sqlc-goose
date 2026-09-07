package api

import (
	"net/http"

	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/reports"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *reports.Service
}

func Register(e *echo.Echo, svc *reports.Service) {
	h := &Handler{svc: svc}
	g := e.Group("/api/v1/reports")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
}

func (h *Handler) Create(c echo.Context) error {
	var req reports.CreateRequest
	if err := c.Bind(&req); err != nil {
		return httpx.NewAPIError(httpx.ErrBadRequest, "invalid body", nil)
	}
	out, err := h.svc.CreateReport(c.Request().Context(), req)
	if err != nil {
		return err
	}
	c.Response().Header().Set(echo.HeaderLocation, "/api/v1/reports/"+out.ReportID.String())
	return c.JSON(http.StatusAccepted, out)
}

func (h *Handler) List(c echo.Context) error {
	out, err := h.svc.ListReports(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) Get(c echo.Context) error {
	out, err := h.svc.GetReport(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}
