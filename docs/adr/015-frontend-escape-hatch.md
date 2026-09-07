# ADR 015 — Escape hatch to a dedicated frontend

## Status

Accepted

## Context

htmx is excellent for CRUD, forms, and async status UIs. Some domains outgrow it.

## Decision

Graduate a screen/domain to a separate frontend when you need substantial client state, drag/drop, canvas/WebGL, offline-first, collaborative editing, or heavy optimistic multi-widget coordination without server round-trips.

The rich client should call `/api/v1/*` against the **same** Go services. Do not force htmx beyond its strengths.

## Consequences

Dual delivery remains valid: keep `/app` for most screens; add a SPA only where justified.
