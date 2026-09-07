# Developer notes

## Migrations

```bash
goose -dir sql/migrations postgres "$DSN" up
goose -dir sql/migrations postgres "$DSN" status
```

## sqlc

```bash
sqlc generate
```

## Multi-instance rate limiting

`RATE_LIMIT_USE_MEMORY=true` is for single-instance/dev only. For multiple app instances, rate-limit at the reverse proxy/edge, or enable a shared store (`RATE_LIMIT_POSTGRES` reserved for future). Document your choice in deployment config.

## Extraction criteria

See ADRs and architecture docs: split DB / extract worker / add Redis only with concrete operational reasons.
