package server

import (
	"fmt"
	"go-echo-server-template/internal/config"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/internal/errors"
	"go-echo-server-template/internal/logger"
	"go-echo-server-template/routes"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// Start initializes and starts the server
func Start() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Initialize logger
	if err := logger.Initialize(os.Getenv("APP_ENV")); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	dbConfig := &config.DatabaseConfig{
		Host: os.Getenv("DB_HOST"),
		Port: func() int {
			p, _ := strconv.Atoi(os.Getenv("DB_PORT"))
			if p == 0 {
				return 5432
			}
			return p
		}(), // Default to 5432 if not set
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := database.Initialize(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// Create queries
	queries := database.New(db)

	// Create and configure server
	e := NewServer()

	// Initialize routes
	InitializeRoutes(e, queries)

	// Start server
	return e.Start(":" + port)
}

// NewServer creates a new Echo server instance with middleware
func NewServer() *echo.Echo {
	e := echo.New()

	// Set custom error handler
	e.HTTPErrorHandler = errors.ErrorHandler

	// Security Middleware
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Request ID Middleware for tracing
	e.Use(middleware.RequestID())

	// Custom Logger Middleware
	e.Use(logger.LoggerMiddleware())

	// Recover Middleware
	e.Use(middleware.Recover())

	// Body Limit Middleware to prevent large payload attacks
	bodyLimit := os.Getenv("BODY_LIMIT")
	if bodyLimit == "" {
		bodyLimit = "2M" // Default 2MB limit
	}
	e.Use(middleware.BodyLimit(bodyLimit))

	// CORS Middleware with secure configuration
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*" // Default allow all in development
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{allowedOrigins},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Rate Limiter with more robust configuration
	rateLimit, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS"))
	if rateLimit == 0 {
		rateLimit = 100 // Default 100 requests per minute
	}
	burstLimit, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_BURST"))
	if burstLimit == 0 {
		burstLimit = 50 // Default burst of 50 requests
	}

	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(rateLimit),
			Burst:     burstLimit,
			ExpiresIn: time.Minute * 1,
		},
	)
	e.Use(middleware.RateLimiter(store))

	return e
}

// InitializeRoutes sets up all the routes for the server
func InitializeRoutes(e *echo.Echo, queries *database.Queries) {
	// Add the routes here
	routes.RegisterTodoRoutes(e, queries)
}
