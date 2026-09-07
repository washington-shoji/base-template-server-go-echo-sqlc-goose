# ADR 003: Dual delivery (REST + htmx)

## Decision

Domains may expose `/api/v1` JSON and/or `/app` HTML/htmx adapters calling the same services in-process.

## Rationale

Server-rendered UI without SPA tax; REST for external/SPA clients. Web adapters never HTTP-call REST.
