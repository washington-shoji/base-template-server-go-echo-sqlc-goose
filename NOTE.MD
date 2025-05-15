# Developer Notes

This file serves as a quick reference for common commands, project-specific notes, and development tips.

## Table of Contents

- [Go Modules](#go-modules)
- [Database](#database)
  - [Connection String](#connection-string)
  - [Goose Migrations](#goose-migrations)
  - [SQLc Queries](#sqlc-queries)
- [Testing](#testing)
  - [Running Tests with Coverage](#running-tests-with-coverage)
- [Docker (Common Commands)](#docker-common-commands)
- [Other Useful Tips](#other-useful-tips)

## Go Modules

Update dependencies and clean up `go.mod` and `go.sum`:

```bash
go mod tidy
```

To vendor dependencies (optional, if your project uses vendoring):

```bash
go mod vendor
```

## Database

### Connection String

Format for PostgreSQL connection string:

```
postgresql://USERNAME:PASSWORD@HOST:PORT/DBNAME?sslmode=disable
```

### Goose Migrations

All Goose commands should be run from the project root or by specifying the directory containing migration files if it's different.
Typically, migrations are in `sql/schema/`.

**Create a new migration:**

- With a sequential timestamp prefix (e.g., `00001_create_users_table.sql`):
  ```bash
  goose -dir sql/schema create create_users_table sql
  ```
- With a sequential timestamp and custom name format (e.g., `00001_CreateUsersTable.sql`):
  ```bash
  goose -dir sql/schema create CreateUsersTable sql
  ```

**Run migrations:**

Replace `<database-string>` with your actual connection string (e.g., `"postgres://user:pass@host:port/db?sslmode=disable"`).

- Apply all pending UP migrations:

  ```bash
  goose -dir sql/schema postgres "<database-string>" up
  ```

- Rollback the last applied migration:

  ```bash
  goose -dir sql/schema postgres "<database-string>" down
  ```

- Rollback all migrations:

  ```bash
  goose -dir sql/schema postgres "<database-string>" reset
  ```

- Show current migration status:
  ```bash
  goose -dir sql/schema postgres "<database-string>" status
  ```

_For more Goose commands (e.g., `up-to`, `down-to`, `up-by-one`), refer to the official Goose documentation or use `goose --help`._

### SQLc Queries

SQLc generates Go code from SQL queries. Ensure `sqlc.yaml` is configured correctly.
Queries are typically located in `sql/queries/`.

**Generate Go code from SQL queries:**

```bash
sqlc generate
```

## Testing

### Running Tests with Coverage

1.  **Generate coverage profile:**

    ```bash
    go test ./... -coverprofile=coverage.out
    ```

2.  **View HTML report:**
    This will open the report in your default browser.

    ```bash
    go tool cover -html=coverage.out -o coverage.html
    ```

3.  **View function coverage summary in terminal (optional):**
    ```bash
    go tool cover -func=coverage.out
    ```

## Docker (Common Commands)

**List running containers:**

```bash
docker ps
```

**View logs for a container:**

```bash
docker logs <container_name_or_id>
```

**Stop a container:**

```bash
docker stop <container_name_or_id>
```

**Remove a container:**

```bash
docker rm <container_name_or_id>
```

## Other Useful Tips

- **(Add your project-specific tips or reminders here)**
- Example: How to connect to a specific staging environment.
- Example: Common gotchas or non-obvious configurations.
