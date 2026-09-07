package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-echo-server-template/internal/platform/httpx"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlashRoundTrip(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	httpx.FlashSuccess(c, "hello")
	flashes := httpx.PeekFlash(c)
	require.Len(t, flashes, 1)
	assert.Equal(t, "success", flashes[0].Type)
	assert.Equal(t, "hello", flashes[0].Message)

	got := httpx.ConsumeFlash(c)
	require.Len(t, got, 1)
	assert.Empty(t, httpx.PeekFlash(c))
}

func TestFormErrors(t *testing.T) {
	errs := httpx.NewFieldErrors(map[string]string{"label": "required"})
	assert.True(t, errs.Has())
	assert.Equal(t, "required", errs.Field("label"))
	assert.Equal(t, "", errs.Field("missing"))
}

func TestIsHTMX(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	assert.False(t, httpx.IsHTMX(c))
	req.Header.Set("HX-Request", "true")
	assert.True(t, httpx.IsHTMX(c))
}
