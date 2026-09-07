# ADR 004: One Postgres with domain schemas

## Decision

One Postgres instance; schemas `todos`, `auth`, `platform`. Physical DB split only when SLO/compliance/ownership require it and cross-domain transactions are already eliminated.
