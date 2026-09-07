package httpx

import (
	"bytes"
	"html"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// Shell is the shared chrome for app/auth layouts. Embed it in page view models.
type Shell struct {
	Title string
	CSRF  string
	Flash []Flash
	User  string
	Nav   string // active nav key: todos|reports|login
}

func (s Shell) shell() Shell { return s }

type shellProvider interface {
	shell() Shell
}

type Renderer struct {
	once sync.Once
	t    *template.Template
	err  error
	root string
	fsys fs.FS // optional; nil = OS filesystem under root
}

func NewRenderer(root string) *Renderer {
	if root == "" {
		root = "web/templates"
	}
	return &Renderer{root: root}
}

// NewRendererFS loads templates from an fs.FS (e.g. embed). Names are paths relative to fsys root.
func NewRendererFS(fsys fs.FS) *Renderer {
	return &Renderer{fsys: fsys, root: "."}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, errDictArgs
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errDictKey
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
		"formatTime": func(t any) string {
			switch v := t.(type) {
			case time.Time:
				if v.IsZero() {
					return ""
				}
				return v.UTC().Format("2006-01-02 15:04 UTC")
			case *time.Time:
				if v == nil || v.IsZero() {
					return ""
				}
				return v.UTC().Format("2006-01-02 15:04 UTC")
			default:
				return ""
			}
		},
		"csrfField": func(token string) template.HTML {
			return template.HTML(`<input type="hidden" name="csrf_token" value="` + html.EscapeString(token) + `">`)
		},
	}
}

var (
	errDictArgs = errString("dict: need even number of args")
	errDictKey  = errString("dict: keys must be strings")
)

type errString string

func (e errString) Error() string { return string(e) }

func (r *Renderer) load() {
	r.once.Do(func() {
		r.t = template.New("").Funcs(templateFuncs())
		if r.fsys != nil {
			r.err = fs.WalkDir(r.fsys, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				if filepath.Ext(path) != ".html" {
					return nil
				}
				b, err := fs.ReadFile(r.fsys, path)
				if err != nil {
					return err
				}
				name := filepath.ToSlash(path)
				_, err = r.t.New(name).Parse(string(b))
				return err
			})
			return
		}
		r.err = filepath.WalkDir(r.root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if filepath.Ext(path) != ".html" {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			name := filepath.ToSlash(path[len(r.root)+1:])
			_, err = r.t.New(name).Parse(string(b))
			return err
		})
	})
}

func (r *Renderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	r.load()
	if r.err != nil {
		return r.err
	}
	return r.t.ExecuteTemplate(w, name, data)
}

func (r *Renderer) execute(name string, data any) (string, error) {
	r.load()
	if r.err != nil {
		return "", r.err
	}
	var buf bytes.Buffer
	if err := r.t.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// DefaultRenderer is set by NewRenderer usage in app; helpers use Echo's renderer when possible.
var defaultRenderer *Renderer

func SetDefaultRenderer(r *Renderer) { defaultRenderer = r }

func rendererFrom(c echo.Context) *Renderer {
	if defaultRenderer != nil {
		return defaultRenderer
	}
	if r, ok := c.Echo().Renderer.(*Renderer); ok {
		return r
	}
	return NewRenderer("web/templates")
}

// IsHTMX reports whether the request was issued by htmx.
func IsHTMX(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

type layoutView struct {
	Shell
	Content template.HTML
	Page    any
}

func extractShell(data any) Shell {
	if sp, ok := data.(shellProvider); ok {
		return sp.shell()
	}
	return Shell{}
}

func renderLayout(c echo.Context, status int, layout, page string, data any) error {
	r := rendererFrom(c)
	body, err := r.execute(page, data)
	if err != nil {
		return err
	}
	view := layoutView{
		Shell:   extractShell(data),
		Content: template.HTML(body),
		Page:    data,
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return r.Render(c.Response(), layout, view, c)
}

// RenderApp wraps a page body in layouts/app.html.
func RenderApp(c echo.Context, page string, data any) error {
	return renderLayout(c, http.StatusOK, "layouts/app.html", page, data)
}

// RenderAppStatus is RenderApp with an explicit status (e.g. 422 validation).
func RenderAppStatus(c echo.Context, status int, page string, data any) error {
	return renderLayout(c, status, "layouts/app.html", page, data)
}

// RenderAuth wraps a page body in layouts/auth.html.
func RenderAuth(c echo.Context, page string, data any) error {
	return renderLayout(c, http.StatusOK, "layouts/auth.html", page, data)
}

// RenderAuthStatus is RenderAuth with an explicit status.
func RenderAuthStatus(c echo.Context, status int, page string, data any) error {
	return renderLayout(c, status, "layouts/auth.html", page, data)
}

// RenderFragment renders a template with no layout (htmx swaps).
func RenderFragment(c echo.Context, name string, data any) error {
	return c.Render(http.StatusOK, name, data)
}

// RenderFragmentStatus renders a fragment with an explicit status.
func RenderFragmentStatus(c echo.Context, status int, name string, data any) error {
	r := rendererFrom(c)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return r.Render(c.Response(), name, data, c)
}
