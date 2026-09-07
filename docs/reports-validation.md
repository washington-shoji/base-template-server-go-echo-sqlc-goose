# Reports domain — architecture validation report

## Implementation summary

Added `internal/reports` as a second, materially different domain:

- User-owned reports with explicit async lifecycle (`pending` / `processing` / `completed` / `failed`)
- Conditional claim, retryable → pending, terminal / attempt-exhausted → failed, stale-processing reclaim
- Async generation via `reports.generate` Postgres jobs + `jobs.AttemptInfo` in handler context
- Enqueue failure leaves **pending**; `ReconcilePending` repairs dual-write orphans (no outbox)
- Cross-domain `TodoSummaryReader` (global todo counts; Todos are not user-owned)
- REST (`202 Accepted` + `Location`) + htmx UI with status polling (stops on terminal)
- Service-level ownership authz; schema `reports.*`, dedicated sqlc package, FakeStore + integration tests

## Lifecycle (copyable reference)

```text
CreateReport → pending
ClaimExecution → processing (or no-op if leased)
  success → completed
  retryable (not last attempt) → pending + RetryableError
  retryable (last attempt) / unsupported type → failed + failure_code
```

| Field | Role |
|-------|------|
| `started_at` | First time processing ever began (debug/UI) |
| `processing_started_at` | Current attempt lease; cleared on pending/completed/failed |
| `failure_code` | Stable code (`UNSUPPORTED_TYPE`, `RETRY_EXHAUSTED`, `INTERNAL`, …) |
| `failure_reason` | Safe user-facing message |

**Not supported:** deletion; admin `failed → pending` retry (document as future).

**Config:** `REPORTS_PROCESSING_LEASE` (5m), `REPORTS_RECONCILE_AFTER` (1m), `REPORTS_RECONCILE_INTERVAL` (30s).

## Files changed

### New / updated Reports code

- `internal/reports/` (domain, store, service, fakes, api, web, postgres)
- `web/templates/reports/`
- `sql/migrations/00004_reports_schema.sql`, `00005_reports_lifecycle.sql`
- `sql/queries/reports/`

### Existing domain changes

- `internal/todos`: `Summary` / `CountTodos` sqlc + Store + Service + FakeStore

### Platform changes

- `internal/platform/jobs`: `AttemptInfo` / context helpers only
- `internal/platform/config`: `ReportsConfig` lease/reconcile knobs

### Composition changes

- `internal/app/app.go`: reports wiring, adapters, job handler, reconcile ticker

### Tests

- Unit: claim races, retryable→pending, last-attempt→failed, enqueue leaves pending, reconcile, htmx poll triggers
- Integration: async completion, ownership, concurrent claim, stale reclaim, reconcile

### Documentation

- `AGENTS.md`, `docs/architecture.md`, `docs/adding-a-domain.md`, this file

## Architecture validation (Section 18)

1. **Platform packages changed:** jobs attempt context + config knobs only (minimal)
2. **Todo files changed:** store/service/fake/sqlc/postgres only (Summary capability)
3. **sqlc multi-domain:** Yes — third block worked cleanly
4. **Goose schemas:** Straightforward `CREATE SCHEMA reports` + lifecycle alter
5. **`internal/app` sufficient:** Yes as composition root
6. **Business logic in app?** No — only adapters + job decode + reconcile loop
7. **Consumer-owned capability:** Easy four-line interface + tiny adapter
8. **Cyclic deps:** Avoided (Reports does not import todos/postgres; app adapts)
9. **Reports know Todo persistence?** No
10. **Cross-domain SQL?** No
11. **Cross-domain TX?** No
12. **Auth REST + web:** Same Principal middleware; ownership in service
13. **Service ownership:** Natural `OwnerUserID == principal.UserID`
14. **Jobs without excess platform coupling:** `Enqueuer` port + AttemptInfo context
15. **Typed job payload:** `GenerateReportJob` + JSON decode in app
16. **Idempotency:** Claim + completed/failed no-op; retryable returns to pending
17. **htmx async:** Polling fragment works without WebSockets
18. **Shared application logic:** REST and web both call `reports.Service`
19. **Errors:** `httpx` sentinels map to JSON and HTML
20. **Operation-scoped contexts:** Preserved (no stored ctx)
21. **Generic abstractions for Reports?** No new frameworks / state-machine packages
22. **Todo-specific abstractions broken?** Global Querier already gone; Summary added narrowly
23. **Dumping ground?** `app.go` grew wiring lines only (acceptable)
24. **Every future domain needs same changes?** Only app registration + sqlc/migration pattern
25. **Add-a-domain docs accurate?** Updated for lifecycle, reconcile, capabilities

## Friction report (A–E)

| Finding | Class |
|---------|--------|
| Explicit `app.New` registration of store/service/routes/jobs/reconcile | **D** acceptable wiring |
| Tiny `todoSummaryAdapter` / `queueEnqueuer` in app | **D** |
| Todos not user-owned → global Summary instead of SummaryForUser | **C** domain reality / **E** avoided fake ownership |
| Create then enqueue without outbox; pending + reconcile | **B** missing reusable outbox (documented); not added |
| Job handler needs system Principal for authenticated Todo Summary | **D** explicit; documented in AGENTS |
| sqlc NullTime/NullString mapping boilerplate in postgres/store | **D** / slight **A** noise but standard sqlc |
| Duplicate Register/NewService patterns vs Todos | **E** premature to abstract at 2 domains |
| Events unused by Reports | **E** correctly unused |

## Template changes made

- Docs/AGENTS: lifecycle reference, AttemptInfo, reconcile vs mark-failed-on-enqueue
- Todos gained `Summary`/`CountTodos` as a reusable narrow export
- Jobs gained minimal attempt context for domain budget exhaustion

## Abstractions deliberately NOT added

- Outbox / distributed TX
- Generic state-machine package
- Auto domain registration / DI
- Shared “BaseService”
- Per-user Todo ownership migration
- Event bus usage for report generation
- Report deletion / admin failed→pending retry

## Remaining reliability gap

Create + enqueue remain two commits; reconciliation repairs orphans. **Outbox remains deferred** until transactional message publication is required across many workflows or reconcile cost becomes operational pain.

## Test results

```text
go test ./internal/... -count=1
# ok

go test ./tests/integration/... -count=1 -timeout 180s
# ok

go build ./...
# ok
```

## Final assessment

```text
PASS
Smart monolith architecture validated by second domain.
Lifecycle semantics tightened into a copyable reference pattern.
```

Adding Reports required mostly new domain code, a small composition-root wiring block, sqlc/migration entries, Todo Summary capability, adapters, and tests — not a platform redesign. Remaining gaps (outbox) are documented limitations, not blockers for the template’s modular-monolith goals.
