# Go Echo SQLc Goose Starter Kit

A base API server built with Go, Echo, SQLc, and Goose. This starter kit provides a solid foundation for building robust and scalable web APIs.

## Table of Contents

- [Go Echo SQLc Goose Starter Kit](#go-echo-sqlc-goose-starter-kit)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Setup](#setup)
  - [Running the Server](#running-the-server)
    - [Local Development](#local-development)
    - [With Docker (TODO)](#with-docker-todo)
  - [Running Tests](#running-tests)
    - [Unit Tests](#unit-tests)
    - [Integration Tests](#integration-tests)
    - [Test Coverage (TODO)](#test-coverage-todo)
  - [Configuration](#configuration)
  - [Project Structure](#project-structure)
  - [Adding New Modules/Domains](#adding-new-modulesdomains)
  - [API Documentation (TODO)](#api-documentation-todo)

## Prerequisites

- **Go**: Version 1.21 or higher. (Check your version with `go version`)
- **Docker**: For running a PostgreSQL database (or use a local/cloud instance).
- **Goose**: For database migrations. Install with `go install github.com/pressly/goose/v3/cmd/goose@latest`.
- **SQLc**: For generating type-safe Go code from SQL. Install with `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`.
- **make** (Optional, but recommended for using Makefile commands if provided in the future).

## Setup

1.  **Clone the repository:**

    ```bash
    git clone <repository-url>
    cd <repository-name>
    ```

2.  **Install dependencies:**

    ```bash
    go mod tidy
    ```

3.  **Set up PostgreSQL:**
    You can use Docker to quickly start a PostgreSQL instance:

    ```bash
    docker run --name my-postgres -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=mydb -p 5432:5432 -d postgres:15-alpine
    ```

    Alternatively, use a local or cloud-based PostgreSQL instance.

4.  **Configure Environment Variables:**
    Copy the `.env.example` file (if provided) to `.env` and update the values, or set the environment variables directly. See the [Configuration](#configuration) section for details.
    Key variables include database connection details (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).

5.  **Run Database Migrations:**
    Ensure your database is running and accessible.
    ```bash
    goose -dir sql/schema postgres "postgres://user:password@localhost:5432/mydb?sslmode=disable" up
    ```
    Replace `user`, `password`, `localhost`, `5432`, and `mydb` with your actual database credentials and schema path.

## Running the Server

### Local Development

To run the server directly using Go:

```bash
go run main.go
```

The server will typically start on the port specified in your configuration (e.g., `8080`).

### With Docker (TODO)

Instructions for building and running the server with Docker will be added here.

## Running Tests

### Unit Tests

To run all unit tests:

```bash
go test ./...
```

To run unit tests for a specific package:

```bash
go test ./path/to/package/...
```

### Integration Tests

Integration tests typically require a running database. Ensure your test database is configured and accessible (see `tests/integration/setup_test.go`).

To run integration tests:

```bash
go test ./tests/integration/...
```

### Test Coverage (TODO)

Instructions for generating and viewing test coverage reports will be added here. Example:

```bash
# go test -coverprofile=coverage.out ./...
# go tool cover -html=coverage.out
```

## Configuration

The application is configured using environment variables. The following variables are commonly used:

| Variable               | Description                                                         | Default Value   |
| ---------------------- | ------------------------------------------------------------------- | --------------- |
| `APP_ENV`              | Application environment (e.g., `development`, `production`, `test`) | `development`   |
| `SERVER_PORT`          | Port for the HTTP server to listen on                               | `8080`          |
| `LOGGER_LEVEL`         | Log level (e.g., `DEBUG`, `INFO`, `WARN`, `ERROR`)                  | `INFO`          |
| `DB_HOST`              | Database host                                                       | `localhost`     |
| `DB_PORT`              | Database port                                                       | `5432`          |
| `DB_USER`              | Database username                                                   | `user`          |
| `DB_PASSWORD`          | Database password                                                   | `password`      |
| `DB_NAME`              | Database name                                                       | `mydb`          |
| `DB_SSL_MODE`          | Database SSL mode (e.g., `disable`, `require`)                      | `disable`       |
| `BODY_LIMIT`           | Maximum request body size (e.g., `2M`)                              | `2M`            |
| `RATE_LIMIT_REQUESTS`  | Max requests per second for rate limiter                            | `10`            |
| `RATE_LIMIT_BURST`     | Max burst requests for rate limiter                                 | `30`            |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins                        | `*` (allow all) |

These are typically loaded from a `.env` file in development (not committed to version control) or set directly in the deployment environment. The `internal/config/config.go` file handles loading these variables.

## Project Structure

```
.
├── cmd/                  # Main applications for the project (if any beyond main.go)
├── internal/             # Private application and library code
│   ├── config/           # Configuration loading and management
│   ├── database/         # Database initialization, SQLC query generation
│   ├── errors/           # Custom error handling
│   ├── logger/           # Logging setup and middleware
│   ├── metrics/          # Prometheus metrics setup
│   ├── models/           # Domain models (can be generated by SQLc or custom)
│   ├── testutils/        # Utilities for testing
├── routes/               # Route definitions
├── handlers/             # HTTP request handlers
├── services/             # Business logic services
├── server/               # Server setup and initialization (Echo instance, middleware)
├── sql/                  # SQL files
│   ├── queries/          # SQL queries for SQLc
│   ├── schema/           # Database migration files (Goose)
├── tests/                # Test files
│   ├── integration/      # Integration tests
│   ├── unit/             # Unit tests (can also be alongside code in packages)
├── vendor/               # Go module dependencies
├── .env.example          # Example environment variables file
├── .gitignore            # Specifies intentionally untracked files that Git should ignore
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── main.go               # Main application entry point
├── prometheus.yml        # Example Prometheus configuration
├── README.md             # This file
└── sqlc.yaml             # SQLc configuration file
```

## Adding New Modules/Domains

Here's a quick guide to adding a new feature or domain (e.g., "Products"):

1.  **Define Database Schema (SQL Migrations):**

    - Create a new SQL migration file in `sql/schema/` (e.g., `YYYYMMDDHHMMSS_create_products_table.sql`).
    - Define your table schema (e.g., `CREATE TABLE products (...)`).
    - Run migrations: `goose -dir sql/schema postgres "YOUR_DB_CONNECTION_STRING" up`

2.  **Write SQL Queries (SQLc):**

    - Create SQL query files in `sql/queries/` (e.g., `product.sql`).
    - Write your SQL queries with SQLc annotations (e.g., `-- name: GetProduct :one`).
    - Refer to SQLc documentation for query syntax.

3.  **Generate Go Code (SQLc):**

    - Run SQLc to generate Go code for your queries:
      ```bash
      sqlc generate
      ```
    - This will generate Go files (e.g., `product.sql.go`) in `internal/database/` (or as configured in `sqlc.yaml`). These files will contain type-safe functions for your queries.

4.  **Define Domain Models (Optional but Recommended):**

    - You might want to define custom domain models in `internal/models/` (e.g., `product.go`) that are separate from the SQLc-generated structs if you need more complex business logic or transformations.

5.  **Create Service Layer:**

    - Create a service interface and implementation in `services/` (e.g., `product_service.go`, `product_service_impl.go`).
    - The service will use the SQLc-generated query functions (e.g., `queries.GetProduct(ctx, id)`).
    - Implement business logic here.

6.  **Create Handler Layer:**

    - Create handlers in `handlers/` (e.g., `product_handler.go`).
    - Handlers will parse requests, call the appropriate service methods, and format responses.
    - Use Echo's context for request/response handling.

7.  **Define Routes:**

    - Define routes in `routes/` (e.g., `product_routes.go`).
    - Map HTTP endpoints to your handlers (e.g., `e.GET("/products/:id", productHandler.GetProduct)`).
    - Register these routes in `server/server.go` within the `InitializeRoutes` function.

8.  **Write Tests:**
    - Add unit tests for your services and handlers.
    - Add integration tests to cover the new endpoints and database interactions.

## API Documentation (TODO)

Information on how to generate and access API documentation (e.g., using Swagger/OpenAPI with tools like `swaggo`) will be added here.

## Disclaimer

THIS SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESSED OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

## License

MIT License

Copyright (c) 2024 

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
