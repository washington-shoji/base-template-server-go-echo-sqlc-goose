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

goose postgres <database-string> up

## SQLC commands

sqlc generate
