package database

import (
	"database/sql"
	"fmt"
	"go-echo-server-template/internal/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// Initialize sets up the database connection and runs migrations
func Initialize(config *config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.ToDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}

// RunMigrations runs the database migrations
func RunMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database for migrations: %v", err)
	}
	defer db.Close()

	// Set the migration directory
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	// Run migrations
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}
