# Develper's Note

## Go commands

go mod tidy

go mod vendor

## DB Connection

postgresql://USERNAME:PASSWORD@HOST:PORT/DBNAME?sslmode=disable

## Goose commands

### Adding a new schema with sequence i.e. 000001...

goose -s create <schema-name> sql

### Generate migration for postgres

<!-- Run the command in the same directory as the schema files -->

<!-- Run migration up for all -->

goose postgres <database-string> up

<!-- Run migration up to a specific version -->

goose <database-string> up-to <migration-version>

<!-- Run migration up a single migration from the current version -->

goose <database-string> up-by-one

<!-- Run migration down for all -->

goose postgres <database-string> down

<!-- Run migration down to a specific version -->

goose <database-string> down-to <migration-version>

## SQLC commands

sqlc generate
