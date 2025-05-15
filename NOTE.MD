# Develper's Note

## Go commands

go mod tidy

go mod vendor

## DB Connection

postgresql://USERNAME:PASSWORD@HOST:PORT/DBNAME?sslmode=disable

## Goose commands

### Adding a new schema with sequence i.e. 000001...

goose -s create <schema-name> sql

### Adding a new schema with sequence and name i.e. 000001[Schema_Name]

goose -s create [Schema-Name] sql

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

## How to Generate and View Coverage Reports

### Run test with coverage

go test ./... -coverprofile=coverage.out

### View HTML Report

go tool cover -html=coverage.out -o coverage.html

### View Text Summary (Optional)

<!-- For a quick command-line summary of coverage percentage per function: -->

go tool cover -func=coverage.out
