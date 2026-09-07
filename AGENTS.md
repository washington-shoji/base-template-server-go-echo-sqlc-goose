# AGENTS.md — Coding agent guide

This repository is a **smart modular monolith** Go template.

## Layout

```
cmd/server          HTTP (+ optional embedded workers)
cmd/worker          Job-only process
internal/platform   Shared infra (config, db, httpx, logging, metrics, authn/authz, jobs, events)
internal/todos      Simple CRUD reference domain
internal/reports    Advanced reference: lifecycle, jobs, cross-domain capability, dual delivery
internal/auth       Auth domain (sessions, bearer tokens, CSRF, web login)
internal/app        Composition root (wiring only)
web/templates       layouts / pages / fragments / components
web/styles          Tailwind source (input.css)
web/static          Generated CSS + vendored JS (embedded in binary)
sql/migrations      Goose migrations (domain schemas)
sql/queries/<dom>   sqlc queries per domain
docs/               Architecture + ADRs
```

## Dependency rules

- Adapters (`api`, `web`) → domain service → domain `Store` → `postgres` (sqlc)
- Domains must not import another domain's `postgres` package
- Cross-domain: **narrow capability interface** owned by the consumer, implemented via the provider service and wired in `internal/app` (see Reports → `TodoSummaryReader`)
- Prefer direct enqueue / service calls over events unless genuine fan-out is needed
- No HTTP types inside domain services (use `platform/httpx` sentinel errors)
- `platform/*` must not import domains

## How to add a domain

1. Migration in `sql/migrations` (prefer `CREATE SCHEMA <domain>`)
2. Queries in `sql/queries/<domain>/`
3. Add sqlc entry in `sqlc.yaml` → `internal/<domain>/postgres`
4. Run `sqlc generate`
5. `domain.go`, `store.go`, `service.go`, optional `api/`, `web/`
6. Wire stores/services/adapters/jobs in `internal/app/app.go`
7. Tests with FakeStore + integration as needed

Reference implementations:

- **Todos** — simple CRUD (+ reference `/app` UI)
- **Reports** — async jobs, ownership authz, cross-domain read, REST + htmx polling UI

Validation write-up: [docs/reports-validation.md](docs/reports-validation.md)  
UI cookbook: [docs/ui.md](docs/ui.md)

## UI (htmx + Tailwind)

- Full page → `web/templates/pages/<domain>/` + `httpx.RenderApp`
- Fragment → `web/templates/fragments/<domain>/` + `httpx.RenderFragment`
- Components → `web/templates/components/` via `{{template "components/..." (dict ...)}}`
- Forms: map service errors to `httpx.FormErrors` in the web adapter; never put HTML in domain errors
- Flash: `httpx.FlashSuccess` / `ConsumeFlash` (cookie)
- Dialogs: native `<dialog>` + `web/static/js/app.js`
- Polling: copy Reports status fragment (`hx-trigger` only while non-terminal)
- After adding Tailwind classes: `make css` and commit `web/static/css/app.css`
- JS allowed only for dialogs, theme toggle, tiny DOM helpers — no SPA/store/form framework
- Escape hatch: dedicated frontend on `/api/v1` when client-side state grows (see ADR 015)

## Jobs

- Register handlers in `internal/app` so `cmd/server` and `cmd/worker` share them
- Use typed payloads at the domain boundary; decode in the registered handler
- Job handler context includes `jobs.AttemptInfo` (current attempt / max) for budget exhaustion
- Create-then-enqueue is two commits: on enqueue failure leave the aggregate **pending**; `ReconcilePending` re-enqueues (no outbox yet)
- Job contexts have no user principal; attach a system principal when calling authenticated capabilities

## Reports lifecycle (reference)

Copy this pattern for other async workflows:

| Status | Meaning |
|--------|---------|
| `pending` | Eligible to run (new, backoff, or recovered). No active owner. |
| `processing` | Worker owns the attempt (`processing_started_at` lease). |
| `completed` | Terminal success (no silent regenerate). |
| `failed` | Terminal failure (`failure_code` + safe `failure_reason`). |

Transitions (conditional SQL): `pending→processing` (claim), `processing→completed`, `processing→pending` (retryable), `processing→failed` (terminal or last job attempt). Stale `processing` may be re-claimed after `REPORTS_PROCESSING_LEASE`. Deletion and `failed→pending` admin retry are not implemented.

REST: `POST` → **202** + `Location: /api/v1/reports/{id}`. Never expose queue attempt/lock fields on the report JSON. htmx polls only while `pending`/`processing`.

## Commands

```bash
make tools && make htmx && make css   # first-time UI tooling
make run                              # go run ./cmd/server
make worker
make css / make css-watch / make dev
goose -dir sql/migrations postgres "$DSN" up
sqlc generate
go test ./internal/...
go test ./tests/integration/...
go mod tidy && go mod vendor
```

## Auth

- `AUTH_DISABLED=true` (default in non-production) injects principal `UserID=anonymous`
- Production forbids `AUTH_DISABLED=true`
- Browser: session cookie + CSRF on `/app`
- API: `Authorization: Bearer <token>` from `POST /api/v1/auth/token`
- Ownership checks belong in the **service**, not only middleware

## Do not

- Add a global God `Querier`
- Put SQL in handlers
- Call REST over HTTP from web adapters
- Query another domain's schema from sqlc
- Cross-domain database transactions / FKs
- Introduce Kafka/Redis/microservices without a concrete need (see `docs/adr/`)
- Invent auto-registration or DI containers for domains
- Add Node/Vite/React/Alpine for ordinary `/app` screens
- Put business logic in templates or arbitrary JS state stores
