package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-echo-server-template/internal/app"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/database"
	"go-echo-server-template/internal/todos"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testEcho  *echo.Echo
	testApp   *app.App
	testDB    *sql.DB
	container testcontainers.Container
	testCfg   *config.Config
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfg.AppEnv = "test"
	cfg.Auth.Disabled = true
	cfg.Jobs.Enabled = true
	cfg.Jobs.EmbedInServer = true
	cfg.Jobs.PollInterval = 200 * time.Millisecond
	cfg.Metrics.Public = true
	cfg.Logging.Level = "ERROR"
	testCfg = cfg

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER": "test", "POSTGRES_PASSWORD": "test", "POSTGRES_DB": "test_db",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	cfg.Database.Host = host
	cfg.Database.Port = port.Int()
	cfg.Database.User = "test"
	cfg.Database.Password = "test"
	cfg.Database.DBName = "test_db"

	workDir, _ := os.Getwd()
	cfg.Database.MigrationsDir = filepath.Join(workDir, "..", "..", "sql", "migrations")

	for i := 0; i < 10; i++ {
		testDB, err = database.Open(cfg.Database)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		fmt.Println(err)
		_ = container.Terminate(ctx)
		os.Exit(1)
	}

	_ = goose.SetDialect("postgres")
	if err := goose.Up(testDB, cfg.Database.MigrationsDir); err != nil {
		fmt.Println("migrate:", err)
		_ = container.Terminate(ctx)
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		fmt.Println(err)
		_ = container.Terminate(ctx)
		os.Exit(1)
	}
	testApp = application
	testEcho = application.Echo
	_ = application.Start(context.Background())

	code := m.Run()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = application.Shutdown(shutdownCtx)
	_ = container.Terminate(ctx)
	os.Exit(code)
}

func TestTodoAPIIntegration(t *testing.T) {
	body, _ := json.Marshal(todos.Params{Label: "Test Todo", Completed: false})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/create-todo", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testEcho.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created todos.Todo
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	t.Run("resource get", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/todos/"+created.TodoID.String(), nil)
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("legacy get", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/todo/"+created.TodoID.String(), nil)
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("update", func(t *testing.T) {
		body, _ := json.Marshal(todos.Params{Label: "Updated", Completed: true})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/todos/"+created.TodoID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/delete-todo/"+created.TodoID.String(), nil)
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		req = httptest.NewRequest(http.MethodGet, "/api/v1/todo/"+created.TodoID.String(), nil)
		rec = httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("validation", func(t *testing.T) {
		body, _ := json.Marshal(todos.Params{Label: ""})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("healthz", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("readyz", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		testEcho.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
