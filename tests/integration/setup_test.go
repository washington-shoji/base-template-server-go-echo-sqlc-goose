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
	appConfig  *config.AppConfig // Store loaded config for tests
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

func setupTestDatabase(ctx context.Context, dbCfg *config.DatabaseConfig) (*sql.DB, error) {
	// Initialize database with retries using the provided dbCfg
	var db *sql.DB
	var err error // Define err once for this function scope
	for i := 0; i < 5; i++ {
		db, err = database.Initialize(dbCfg)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database after retries: %v", err)
	}

	// Run migrations
	if err = goose.SetDialect("postgres"); err != nil { // Assign to existing err
		return nil, fmt.Errorf("failed to set dialect: %v", err)
	}

	// Run migrations from sql/schema directory using absolute path
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(workDir, "..", "..")
	if err = goose.Up(db, filepath.Join(projectRoot, "sql", "schema")); err != nil { // Assign to existing err
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return db, nil
}

func TestMain(m *testing.M) {
	var err error // Define err once for TestMain scope
	appConfig, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config for tests: %v\n", err)
		os.Exit(1)
	}
	appConfig.AppEnv = "test"
	appConfig.Logger.Level = "DEBUG"

	if err = logger.Initialize(appConfig.AppEnv, appConfig.Logger.Level); err != nil { // Assign to existing err
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	container, err = setupTestContainer(ctx)
	if err != nil {
		fmt.Printf("Failed to setup test container: %v\n", err)
		os.Exit(1)
	}

	// Override database config for test container
	mappedPort, portErr := container.MappedPort(ctx, "5432")
	if portErr != nil {
		fmt.Printf("Failed to get container mapped port: %v\n", portErr)
		container.Terminate(ctx)
		os.Exit(1)
	}
	host, hostErr := container.Host(ctx)
	if hostErr != nil {
		fmt.Printf("Failed to get container host: %v\n", hostErr)
		container.Terminate(ctx)
		os.Exit(1)
	}
	appConfig.Database.Host = host
	appConfig.Database.Port = mappedPort.Int()
	appConfig.Database.User = "test"
	appConfig.Database.Password = "test"
	appConfig.Database.DBName = "test_db"

	testDB, err = setupTestDatabase(ctx, &appConfig.Database)
	if err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		if container != nil {
			container.Terminate(ctx)
		}
		os.Exit(1)
	}

	queries = database.New(testDB)
	testServer = server.NewServer(appConfig)
	server.InitializeRoutes(testServer, queries)

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}
	if container != nil {
		container.Terminate(ctx)
	}
	os.Exit(code)
}
