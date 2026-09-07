package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/config"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Middleware attaches a Principal from session cookie or bearer token.
func Middleware(svc *Service, cfg config.AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			if cfg.Disabled {
				ctx = authn.WithPrincipal(ctx, authn.Principal{
					UserID: "anonymous", Roles: []string{"user"}, AuthMethod: "disabled",
				})
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}

			p := authn.Anonymous()
			authz := c.Request().Header.Get("Authorization")
			if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
				token := strings.TrimSpace(authz[7:])
				if principal, err := svc.PrincipalFromBearer(ctx, token); err == nil {
					p = principal
				}
			} else if cookie, err := c.Cookie(cfg.SessionCookie); err == nil && cookie.Value != "" {
				if sid, err := uuid.Parse(cookie.Value); err == nil {
					if principal, err := svc.PrincipalFromSession(ctx, sid); err == nil {
						p = principal
					}
				}
			}

			ctx = authn.WithPrincipal(ctx, p)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// CSRFMiddleware validates CSRF token for unsafe /app methods when auth is enabled.
func CSRFMiddleware(cfg config.AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Disabled {
				return next(c)
			}
			path := c.Request().URL.Path
			if !strings.HasPrefix(path, "/app") {
				return next(c)
			}
			method := c.Request().Method
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				ensureCSRFCookie(c, cfg)
				return next(c)
			}
			cookie, err := c.Cookie(cfg.CSRFCookieName)
			header := c.Request().Header.Get("X-CSRF-Token")
			if header == "" {
				header = c.FormValue("csrf_token")
			}
			if err != nil || cookie.Value == "" || header == "" || cookie.Value != header {
				return c.JSON(http.StatusForbidden, map[string]string{"message": "csrf validation failed"})
			}
			return next(c)
		}
	}
}

func ensureCSRFCookie(c echo.Context, cfg config.AuthConfig) {
	if _, err := c.Cookie(cfg.CSRFCookieName); err == nil {
		return
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	c.SetCookie(&http.Cookie{
		Name:     cfg.CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
}

func NewCSRFToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
