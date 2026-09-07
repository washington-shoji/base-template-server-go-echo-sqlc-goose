# UI cookbook (htmx + Tailwind)

Server-rendered HTML is the default application UI. REST stays under `/api/v1/*`.

## Where things go

| Need | Location |
|------|----------|
| Full page body | `web/templates/pages/<domain>/` |
| htmx fragment | `web/templates/fragments/<domain>/` |
| Layout chrome | `web/templates/layouts/{app,auth}.html` |
| Reusable UI | `web/templates/components/` |
| CSS source | `web/styles/input.css` |
| Generated CSS | `web/static/css/app.css` (committed) |
| JS | `web/static/js/{htmx.min.js,app.js,theme-boot.js}` |

## Render helpers

```go
httpx.RenderApp(c, "pages/todos/list.html", data)       // layouts/app.html
httpx.RenderAuth(c, "pages/auth/login.html", data)
httpx.RenderFragment(c, "fragments/todos/list.html", data)

if httpx.IsHTMX(c) { /* fragment */ } else { /* full page */ }
```

Embed `httpx.Shell` in page view models (`Title`, `CSRF`, `Flash`, `User`, `Nav`).

## Forms and validation

1. Bind form values in the **web adapter**
2. Call the **service** (transport-neutral errors)
3. On failure: map to `httpx.FormErrors` and re-render (200/422)
4. On success: `httpx.FlashSuccess` + redirect, or return an htmx fragment

Preserve submitted values. Use labels + `aria-describedby` for field errors. CSRF via `components/csrf-field.html`.

## Flash

```go
httpx.FlashSuccess(c, "Saved")
httpx.ConsumeFlash(c) // in page Shell for full-page renders
```

Cookie-backed, one-shot. Prefer server-rendered alerts over toast frameworks.

## Buttons / dialogs

```html
{{template "components/button.html" (dict "Label" "Save" "Variant" "primary" "Type" "submit")}}
```

Dialogs: native `<dialog>` + `data-dialog-open` / `data-dialog-close` / `data-dialog-confirm` handled in `app.js`.

## Polling (Reports)

Status fragment self-attaches `hx-trigger="every 2s"` only while `pending`/`processing`. Use `aria-live="polite"` and `components/status-badge.html`.

## Tailwind

```bash
make tools      # download standalone CLI → .tools/
make css        # build minified app.css
make css-watch  # watch mode
make css-check  # CI drift check
make htmx       # vendor htmx
make dev        # css-watch + server
```

No Node/npm. After editing templates/classes, run `make css` and commit `web/static/css/app.css`.

## JavaScript policy

Allowed in `app.js`: dialog open/close, theme toggle, flash dismiss, tiny DOM helpers.

Do **not** add client routing, stores, form frameworks, or Alpine by default.

## Escape hatch

If a screen needs growing client-side state, drag/drop, canvas, offline, or collaborative editing, build a dedicated frontend against `/api/v1` using the same Go services.

## References

- **Todos** — CRUD UI (`/app/todos`)
- **Reports** — async status UI (`/app/reports`)
