# ADR 001: Modular monolith

## Decision

Ship one deployable Go application with domain packages, not microservices.

## Rationale

Operational simplicity; strong internal boundaries for later extraction.

## Consequences

Shared process/DB initially; extract worker/service only with concrete criteria (see architecture docs).
