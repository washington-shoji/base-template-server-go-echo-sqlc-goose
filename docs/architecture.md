# Architecture overview

## Shape

One Go application with domain packages under `internal/`, dual delivery adapters (REST + HTML/htmx), shared Postgres with per-domain schemas, in-process events, and a Postgres-backed job queue.

```
/app/*     HTML + htmx  ──┐
/api/v1/*  JSON REST    ──┼── domain services ── Store ── sqlc ── Postgres schemas
workers                 ──┘
```

## Request flow

1. Echo middleware (metrics, secure headers, request ID → context, logging, recover, body limit, CORS, rate limit)
2. Auth middleware attaches `authn.Principal`
3. CSRF on mutating `/app` routes when auth enabled
4. Adapter (`api` or `web`) calls domain service with request context
5. Service validates, authorizes (including ownership), uses Store; may enqueue jobs or publish events
6. Errors are domain sentinels; `httpx.ErrorHandler` maps to JSON or HTML

## Domain ownership

| Schema | Owner |
|--------|--------|
| `todos.*` | `internal/todos` |
| `auth.*` | `internal/auth` |
| `reports.*` | `internal/reports` |
| `platform.*` | jobs / platform |

## Reference domains

- **Todos** — simple authenticated CRUD (not user-owned rows); `/app/todos` is the CRUD UI reference
- **Reports** — user-owned async workflow; `todo-summary` reads Todos via `TodoSummaryReader`; REST **202** + htmx polling UI with status badges

UI foundation: [docs/ui.md](ui.md) (Tailwind standalone, layouts/components, forms/flash, JS policy).

## Persistence

- Goose: `sql/migrations` (not auto-run on boot)
- sqlc: one package per domain under `internal/<domain>/postgres`
- Transactions: single-domain only
- No cross-schema FKs

## Jobs and events

- Sync `events.Bus` for optional in-process fan-out (Todos audit uses it; Reports does not)
- `platform.jobs` + `SKIP LOCKED` worker; typed payloads at domain boundary
- Handler context carries `jobs.AttemptInfo` so domains can fail terminal on last attempt without reading `platform.jobs`
- Create-then-enqueue without outbox: enqueue failure → leave aggregate **pending**; Reports `ReconcilePending` ticker re-enqueues orphans (`REPORTS_RECONCILE_*`)
- Embedded in `cmd/server` when `JOBS_EMBED_IN_SERVER=true`, or `cmd/worker`

## Ops endpoints

- `GET /healthz` — liveness
- `GET /readyz` — DB (+ jobs store) readiness
- `GET /metrics` — Prometheus (public only when `METRICS_PUBLIC=true`; otherwise basic auth)
- Legacy: `GET /health`, `GET /`

## Todo API compatibility

Canonical: `POST/GET /api/v1/todos`, `GET/PUT/DELETE /api/v1/todos/:id`

Legacy aliases still registered: `/create-todo`, `/update-todo/:todo-id`, `/delete-todo/:todo-id`, `/todo/:todo-id`

## Reports API and lifecycle

- `POST /api/v1/reports` → **202 Accepted**, `Location: /api/v1/reports/{id}`, body includes `status` / optional `failure_code` / safe `failure_reason`
- `GET /api/v1/reports`, `GET /api/v1/reports/:id`
- Web: `/app/reports`, `/app/reports/:id` (htmx status polling stops on terminal statuses)
- States: `pending` → `processing` → `completed` | `pending` (retryable) | `failed` (terminal / retry exhausted)
- Lease: `processing_started_at` + `REPORTS_PROCESSING_LEASE` (default 5m); `started_at` is first-touch only
- Deletion unsupported; admin regenerate (`failed→pending`) deferred
- Outbox still deferred until many workflows need atomic enqueue
