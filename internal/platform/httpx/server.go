package httpx

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/database"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/platform/metrics"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

type ReadyCheck func(ctx context.Context) error

func NewEcho(cfg *config.Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = ErrorHandler

	e.Use(metrics.Middleware)
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'; style-src 'self'; script-src 'self'; img-src 'self' data:; font-src 'self'; frame-ancestors 'self'; base-uri 'self'; form-action 'self'",
	}))
	e.Use(middleware.RequestID())
	e.Use(logging.Middleware())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	allowCredentials := true
	if len(cfg.CORS.AllowedOrigins) == 1 && cfg.CORS.AllowedOrigins[0] == "*" {
		// Browsers reject AllowCredentials with wildcard origins.
		allowCredentials = false
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions, http.MethodPatch},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowCredentials,
		MaxAge:           300,
	}))

	if cfg.RateLimit.UseMemory {
		store := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(cfg.RateLimit.Requests),
			Burst:     cfg.RateLimit.Burst,
			ExpiresIn: time.Minute,
		})
		e.Use(middleware.RateLimiter(store))
	}

	return e
}

// MountStaticFS serves /static from an fs.FS root that already points at the static directory.
func MountStaticFS(e *echo.Echo, fsys fs.FS) {
	e.StaticFS("/static", fsys)
}

// MountStaticDir serves /static from a filesystem directory (dev / tests).
func MountStaticDir(e *echo.Echo, dir string) {
	e.Static("/static", dir)
}

func RegisterOps(e *echo.Echo, cfg *config.Config, db *sql.DB, extraReady ...ReadyCheck) {
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	// Legacy aliases
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "message": "Server is healthy"})
	})
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "message": "Server is healthy"})
	})

	e.GET("/readyz", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := database.Ready(ctx, db); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "error": "database"})
		}
		for _, check := range extraReady {
			if err := check(ctx); err != nil {
				return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			}
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	if cfg.Metrics.Enabled {
		h := promhttp.HandlerFor(metrics.AppRegistry, promhttp.HandlerOpts{})
		metrics.MustRegister()
		e.GET("/metrics", func(c echo.Context) error {
			if !cfg.Metrics.Public {
				user, pass, ok := c.Request().BasicAuth()
				if !ok || user != cfg.Metrics.BasicAuthUser || pass != cfg.Metrics.BasicAuthPass {
					c.Response().Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
					return c.NoContent(http.StatusUnauthorized)
				}
			}
			h.ServeHTTP(c.Response(), c.Request())
			return nil
		})
	}
}
