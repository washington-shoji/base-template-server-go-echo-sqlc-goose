package config

import (
	"database/sql"
	"fmt"
	"go-echo-server-template/internal/database"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDatabaseConfig creates a new database configuration from environment variables
func NewDatabaseConfig() (*DatabaseConfig, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL could not be found in the process environment")
	}

	// Default values
	config := &DatabaseConfig{
		URL:             dbURL,
		MaxOpenConns:    25,              // Default max open connections
		MaxIdleConns:    25,              // Default max idle connections
		ConnMaxLifetime: time.Minute * 5, // Default connection max lifetime
	}

	// Override defaults with environment variables if they exist
	if maxOpen := os.Getenv("DB_MAX_OPEN_CONNS"); maxOpen != "" {
		if val, err := time.ParseDuration(maxOpen); err == nil {
			config.MaxOpenConns = int(val)
		}
	}

	if maxIdle := os.Getenv("DB_MAX_IDLE_CONNS"); maxIdle != "" {
		if val, err := time.ParseDuration(maxIdle); err == nil {
			config.MaxIdleConns = int(val)
		}
	}

	if maxLifetime := os.Getenv("DB_CONN_MAX_LIFETIME"); maxLifetime != "" {
		if val, err := time.ParseDuration(maxLifetime); err == nil {
			config.ConnMaxLifetime = val
		}
	}

	return config, nil
}

// InitializeDatabase creates and configures a new database connection
func InitializeDatabase(config *DatabaseConfig) (*database.Queries, error) {
	// Create the database connection
	conn, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(config.MaxOpenConns)
	conn.SetMaxIdleConns(config.MaxIdleConns)
	conn.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("cannot ping the database: %w", err)
	}

	// Create and return the database queries instance
	return database.New(conn), nil
}
