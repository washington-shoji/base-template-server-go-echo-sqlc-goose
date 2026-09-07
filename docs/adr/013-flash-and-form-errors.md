# ADR 013 — Cookie flash + FormErrors in web adapters

## Status

Accepted

## Context

Domain errors must stay transport-neutral. Browser UX needs flash messages and field errors.

## Decision

- `httpx.Flash` via one-shot cookie (+ request context for same-request reads)
- `httpx.FormErrors` mapped in web adapters only
- No HTML strings inside domain errors

## Consequences

REST and web share services; only web knows about flash/form presentation.
