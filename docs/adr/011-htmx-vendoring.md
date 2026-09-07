# ADR 011 — htmx 2.0.4 vendored locally

## Status

Accepted

## Context

htmx was loaded from unpkg CDN, forcing CSP exceptions and non-deterministic runtime dependency.

## Decision

Pin **htmx 2.0.4**, vendor as `web/static/js/htmx.min.js` via `make htmx`, checksum in `tools/versions.env`.

Upgrade only for security fixes or a required feature; bump pin + checksum + re-vendor.

## Consequences

Offline/deterministic builds; CSP `script-src 'self'`.
