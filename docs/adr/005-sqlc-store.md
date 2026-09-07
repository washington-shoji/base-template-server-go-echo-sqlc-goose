# ADR 005: sqlc per-domain Store

## Decision

Each domain owns sqlc output under `internal/<domain>/postgres` behind a domain `Store` interface shaped like sqlc methods. No global composite Querier.
