package httpx

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	flashCookieName = "flash_messages"
	flashContextKey = "httpx.flash"
)

// Flash is a one-shot UI message (cookie-backed).
type Flash struct {
	Type    string // success|error|info|warning
	Message string
}

// SetFlash stores flashes for the next response (cookie) and current context.
func SetFlash(c echo.Context, flashes ...Flash) {
	if len(flashes) == 0 {
		return
	}
	all := append(PeekFlash(c), flashes...)
	c.Set(flashContextKey, all)
	raw, err := json.Marshal(all)
	if err != nil {
		return
	}
	c.SetCookie(&http.Cookie{
		Name:     flashCookieName,
		Value:    base64.RawURLEncoding.EncodeToString(raw),
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// PeekFlash returns flashes without clearing them.
func PeekFlash(c echo.Context) []Flash {
	if v := c.Get(flashContextKey); v != nil {
		if flashes, ok := v.([]Flash); ok {
			return flashes
		}
	}
	cookie, err := c.Cookie(flashCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil
	}
	var out []Flash
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	c.Set(flashContextKey, out)
	return out
}

// ConsumeFlash returns flashes and clears cookie + context.
func ConsumeFlash(c echo.Context) []Flash {
	out := PeekFlash(c)
	c.Set(flashContextKey, []Flash(nil))
	clearFlash(c)
	return out
}

func clearFlash(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// FlashSuccess is a convenience for SetFlash success messages.
func FlashSuccess(c echo.Context, msg string) {
	SetFlash(c, Flash{Type: "success", Message: msg})
}

// FlashError is a convenience for SetFlash error messages.
func FlashError(c echo.Context, msg string) {
	SetFlash(c, Flash{Type: "error", Message: msg})
}

// FlashInfo is a convenience for SetFlash info messages.
func FlashInfo(c echo.Context, msg string) {
	SetFlash(c, Flash{Type: "info", Message: msg})
}
