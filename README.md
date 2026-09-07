# Go Modular Monolith Template

Smart modular monolith starter: **Go 1.27 + Echo + sqlc + Goose + Postgres**.

Domains own their services, SQL, and adapters. Deliver HTML/htmx and/or REST from the same application layer. Background jobs use Postgres; auth supports cookies and bearer tokens.

## Prerequisites

- Go 1.27.1+
- Docker (Postgres for local/dev and integration tests)
- [Goose](https://github.com/pressly/goose) and [sqlc](https://docs.sqlc.dev/)

## Quick start

```bash
cp .env.example .env
# start Postgres, then:
goose -dir sql/migrations postgres "postgres://postgres:password@localhost:5432/app_db?sslmode=disable" up
make tools && make htmx && make css   # once (Tailwind CLI + htmx + CSS)
make run                              # or: go run ./cmd/server
```

- API: `http://localhost:8080/api/v1/todos`
- App UI: `http://localhost:8080/app/todos` · Reports: `/app/reports`
- Health: `/healthz` · Ready: `/readyz` · Metrics: `/metrics`

UI stack: **html/template + htmx + Tailwind standalone** (no Node). See [docs/ui.md](docs/ui.md).

Default: `AUTH_DISABLED=true` for local demos. Set `AUTH_DISABLED=false` and register via `/api/v1/auth/register` for real auth.

Worker-only process:

```bash
go run ./cmd/worker
```

## Tests

```bash
go test ./internal/...
go test ./tests/integration/...   # requires Docker
make css-check                    # fail if committed CSS is stale
```

## Project structure

See [AGENTS.md](AGENTS.md) and [docs/architecture.md](docs/architecture.md).

```
cmd/server, cmd/worker
internal/platform/*     shared infrastructure
internal/todos          reference domain
internal/auth           sessions / tokens / CSRF
internal/app            wiring
web/templates, web/static
sql/migrations, sql/queries
docs/, docs/adr/
```

## Adding a domain

See [docs/adding-a-domain.md](docs/adding-a-domain.md).

## Configuration

Typed config in `internal/platform/config`. Copy `.env.example`. Important flags:

| Variable | Notes |
|----------|--------|
| `AUTH_DISABLED` | Demo mode; forbidden in production |
| `METRICS_PUBLIC` | If false, require basic auth on `/metrics` |
| `RATE_LIMIT_USE_MEMORY` | Process-local; use edge/Postgres for multi-instance |
| `JOBS_EMBED_IN_SERVER` | Run workers inside `cmd/server` |

## Security note

This is a **template**, not a turnkey production deployment. Review auth, metrics exposure, CORS, TLS, and secrets before production use.

## License

MIT
