# ADR 010 — UI foundation: html/template + htmx + Tailwind standalone

## Status

Accepted

## Context

The template delivers `/app` HTML alongside `/api/v1` REST. We need a polished UI kit without adopting a SPA framework or Node toolchain.

## Decision

- Server-rendered `html/template` + htmx 2.0.4 (vendored)
- Tailwind CSS via **standalone CLI** (no npm)
- Generated `web/static/css/app.css` is committed; CLI binary is not (`.tools/`)
- Layouts / pages / fragments / components under `web/templates/`
- Tiny vanilla JS only (`app.js`, `theme-boot.js`)

## Consequences

Agents can copy Todos/Reports patterns without frontend project setup. CSS rebuild requires `make css`. CSP can stay `'self'` for scripts/styles.
