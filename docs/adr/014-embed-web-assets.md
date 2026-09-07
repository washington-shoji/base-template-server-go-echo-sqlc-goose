# ADR 014 — Embed web templates and static assets

## Status

Accepted

## Context

Filesystem `web/` paths break when the binary is run from another working directory.

## Decision

Embed `web/templates` and `web/static` via `web.Files` (`embed.FS`). `internal/app` mounts templates/static from the embed. Unit tests may still use disk `NewRenderer("../../../web/templates")` for convenience.

## Consequences

Single-binary deploys include UI assets. Template edits require rebuild to appear in `cmd/server` (use `make css-watch` + rebuild, or temporarily point at disk for live HTML iteration).
