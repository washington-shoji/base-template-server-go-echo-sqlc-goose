# ADR 006: Events and jobs

## Decision

Synchronous in-process event bus first; Postgres `SKIP LOCKED` job queue before Redis/Kafka. Outbox when multi-process durability requires it.
