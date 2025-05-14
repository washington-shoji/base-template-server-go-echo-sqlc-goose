package integration

import (
	"context"
	"database/sql"
	"fmt"
	"go-echo-server-template/internal/config"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/internal/logger"
	"go-echo-server-template/server"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testServer *echo.Echo
	testDB     *sql.DB
	queries    *database.Queries
	container  testcontainers.Container
)

func setupTestContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	return container, nil
}

func setupTestDatabase(ctx context.Context, container testcontainers.Container) (*sql.DB, error) {
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container external port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}

	dbConfig := &config.DatabaseConfig{
		Host:            host,
		Port:            mappedPort.Int(),
		User:            "test",
		Password:        "test",
		DBName:          "test_db",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	}

	// Initialize database with retries
	var db *sql.DB
	for i := 0; i < 5; i++ {
		db, err = database.Initialize(dbConfig)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database after retries: %v", err)
	}

	// Run migrations
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set dialect: %v", err)
	}

	// Run migrations from sql/schema directory using absolute path
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}
	// Go up one directory level since we're in tests/integration
	projectRoot := filepath.Join(workDir, "..", "..")
	if err := goose.Up(db, filepath.Join(projectRoot, "sql", "schema")); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return db, nil
}

func TestMain(m *testing.M) {
	// Initialize logger
	if err := logger.Initialize("test"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Start PostgreSQL container
	var err error
	container, err = setupTestContainer(ctx)
	if err != nil {
		fmt.Printf("Failed to setup test container: %v\n", err)
		os.Exit(1)
	}

	// Setup test database
	testDB, err = setupTestDatabase(ctx, container)
	if err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		container.Terminate(ctx)
		os.Exit(1)
	}

	// Initialize queries
	queries = database.New(testDB)

	// Setup test server
	testServer = server.NewServer()
	server.InitializeRoutes(testServer, queries)

	// Run tests
	code := m.Run()

	// Cleanup
	if testDB != nil {
		testDB.Close()
	}
	if container != nil {
		container.Terminate(ctx)
	}

	os.Exit(code)
}
