package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-echo-server-template/internal/platform/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.ToDSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// RunMigrations applies Goose migrations from cfg.MigrationsDir.
// Not called automatically on application start.
func RunMigrations(cfg config.DatabaseConfig) error {
	db, err := sql.Open("postgres", cfg.ToDSN())
	if err != nil {
		return fmt.Errorf("open database for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	dir := cfg.MigrationsDir
	if dir == "" {
		dir = "sql/migrations"
	}
	return goose.Up(db, dir)
}

func Ready(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}
