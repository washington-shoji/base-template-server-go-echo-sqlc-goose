# Adding a new domain

Two references:

| Domain | Use when learning |
|--------|-------------------|
| [`internal/todos`](../internal/todos) | Simple CRUD, REST + htmx |
| [`internal/reports`](../internal/reports) | Lifecycle, jobs, ownership, cross-domain capability |

## Steps

1. **Migration** — `sql/migrations/NNNNN_<domain>_schema.sql` with `CREATE SCHEMA` + tables (no FKs into other domains).
2. **Queries** — `sql/queries/<domain>/*.sql` with sqlc annotations.
3. **sqlc.yaml** — add a `sql:` block with `out: internal/<domain>/postgres`.
4. **Generate** — `sqlc generate`.
5. **Store adapter** — `internal/<domain>/postgres/store.go` mapping sqlc types to domain types.
6. **Domain** — `domain.go`, `store.go` (interface), `service.go` (operation-scoped `ctx`; no stored context).
7. **Adapters** — `api/Register`, optionally `web/Register`.
8. **Wire** — `internal/app/app.go` (stores, services, capability adapters, job handlers, routes).
9. **Tests** — FakeStore unit tests; integration as needed.

## REST

Register under `/api/v1/<resource>`. Domains need not be CRUD-shaped (Reports only creates/lists/gets).

## htmx

Register under `/app/...`. Call the same service; templates in `web/templates/pages` and `web/templates/fragments`. Use `httpx.IsHTMX` + `RenderApp` / `RenderFragment`. See [docs/ui.md](ui.md).

## Cross-domain

Define a **narrow interface in the consuming domain**. Implement it by adapting the provider service in `internal/app`. Do not import `*/postgres` across domains. Do not join foreign schemas in SQL.

Example: Reports defines `TodoSummaryReader`; app wires `todoSummaryAdapter` over `todos.Service.Summary`. Todos are **not** user-owned today, so the summary is global — do not invent ownership to force a `SummaryForUser`.

## Async work

Enqueue via a small `Enqueuer` port or `jobs.Queue` from the composition root. Register job handlers in `app.New`. Prefer jobs over events for durable async work.

### Lifecycle pattern (Reports)

1. Insert aggregate as **`pending`**, then enqueue. If enqueue fails, **leave pending** and rely on reconciliation (do not mark failed).
2. Worker **claims** with conditional SQL (`pending` or stale `processing` → `processing` + lease timestamp). Concurrent claim → no-op.
3. Classify failures: **retryable** → `processing → pending` and return error so the job backs off; on last `jobs.AttemptInfo` → `failed` with `RETRY_EXHAUSTED`. Terminal domain errors → `failed` + `failure_code`, return `nil` to complete the job.
4. Success → `completed` (terminal). Do not silently regenerate.
5. Wire a reconcile ticker when jobs are enabled to re-enqueue old pending rows.
6. Expose status / safe failure fields on the API; never queue lock/attempt columns. htmx should poll only while non-terminal.

Deletion and admin retry from `failed` are optional product features — document if deferred.
