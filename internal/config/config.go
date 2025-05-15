package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// AppConfig holds all application configurations
type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	// Add other configs here, e.g., CORS, RateLimit, BodyLimit
	CORSOptions      CORSOptions
	RateLimitOptions RateLimitOptions
	BodyLimit        string
	AppEnv           string // e.g., "development", "staging", "production"
}

// ServerConfig holds server-specific configurations
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database connection configurations
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// ToDSN converts DatabaseConfig to a DSN string for PostgreSQL
func (dc *DatabaseConfig) ToDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dc.Host, dc.Port, dc.User, dc.Password, dc.DBName, dc.SSLMode)
}

// LoggerConfig holds logger configurations
type LoggerConfig struct {
	Level string // e.g., "DEBUG", "INFO", "ERROR"
}

// CORSOptions holds CORS middleware configurations
type CORSOptions struct {
	AllowedOrigins []string
}

// RateLimitOptions holds rate limiter configurations
type RateLimitOptions struct {
	Requests int
	Burst    int
}

// LoadConfig loads application configuration from environment variables
// and .env file (if present) with default values.
func LoadConfig() (*AppConfig, error) {
	// Load .env file, ignore error if it doesn't exist (common in prod)
	_ = godotenv.Load()

	cfg := &AppConfig{
		// Set default values
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "password",
			DBName:          "app_db",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Logger: LoggerConfig{
			Level: "INFO", // Default to INFO, can be overridden by APP_ENV
		},
		CORSOptions: CORSOptions{
			AllowedOrigins: []string{"*"}, // Default allow all
		},
		RateLimitOptions: RateLimitOptions{
			Requests: 100, // Default 100 requests per minute
			Burst:    50,  // Default burst of 50 requests
		},
		BodyLimit: "2M",          // Default 2MB limit
		AppEnv:    "development", // Default environment
	}

	// Override with environment variables if they are set
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPortStr := os.Getenv("DB_PORT"); dbPortStr != "" {
		if p, err := strconv.Atoi(dbPortStr); err == nil {
			cfg.Database.Port = p
		}
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.Database.Password = dbPass
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.DBName = dbName
	}
	if dbSSLMode := os.Getenv("DB_SSLMODE"); dbSSLMode != "" {
		cfg.Database.SSLMode = dbSSLMode
	}

	if appEnv := os.Getenv("APP_ENV"); appEnv != "" {
		cfg.AppEnv = appEnv
		// Adjust logger level based on APP_ENV if not explicitly set by LOG_LEVEL
		if os.Getenv("LOG_LEVEL") == "" { // Only if LOG_LEVEL is not set
			switch appEnv {
			case "production":
				cfg.Logger.Level = "INFO"
			case "staging":
				cfg.Logger.Level = "DEBUG"
			default: // development or other
				cfg.Logger.Level = "DEBUG"
			}
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logger.Level = logLevel // Explicit LOG_LEVEL overrides APP_ENV convention
	}

	if allowedOrigins := os.Getenv("ALLOWED_ORIGINS"); allowedOrigins != "" {
		// Assuming comma-separated string for multiple origins
		// In a real scenario, you might want more robust parsing or a specific format
		cfg.CORSOptions.AllowedOrigins = []string{allowedOrigins} // Simplistic for now, might need splitting
	}

	if rateLimitRequestsStr := os.Getenv("RATE_LIMIT_REQUESTS"); rateLimitRequestsStr != "" {
		if r, err := strconv.Atoi(rateLimitRequestsStr); err == nil {
			cfg.RateLimitOptions.Requests = r
		}
	}
	if rateLimitBurstStr := os.Getenv("RATE_LIMIT_BURST"); rateLimitBurstStr != "" {
		if b, err := strconv.Atoi(rateLimitBurstStr); err == nil {
			cfg.RateLimitOptions.Burst = b
		}
	}

	if bodyLimit := os.Getenv("BODY_LIMIT"); bodyLimit != "" {
		cfg.BodyLimit = bodyLimit
	}

	return cfg, nil
}
