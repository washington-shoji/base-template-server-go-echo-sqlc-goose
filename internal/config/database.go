package config

import (
	"fmt"
	"time"
)

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	// Connection settings
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (c *DatabaseConfig) ToDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// NewDefaultConfig returns a default database configuration
func NewDefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "todo_db",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	}
}
