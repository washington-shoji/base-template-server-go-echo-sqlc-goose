package server

import (
	"context"
	"database/sql"
	"fmt"
	"go-echo-server-template/internal/config"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/internal/errors"
	"go-echo-server-template/internal/logger"
	"go-echo-server-template/internal/metrics"
	"go-echo-server-template/routes"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

// Start initializes and starts the server
func Start() (*echo.Echo, *sql.DB, error) {
	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use a basic fmt.Printf here as logger might not be initialized or might depend on config
		fmt.Printf("Error loading configuration: %v\n", err)
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger using the loaded configuration
	if err := logger.Initialize(cfg.AppEnv, cfg.Logger.Level); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize database using the loaded configuration
	// Note: We are passing a pointer to cfg.Database as DatabaseConfig might be large
	db, err := database.Initialize(&cfg.Database) // Pass the DatabaseConfig from AppConfig
	if err != nil {
		logger.WithContext(context.Background(), "startup").Fatal("Failed to initialize database", err, nil)
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err) // For main.go panic
	}

	// Create queries
	queries := database.New(db)

	// Create and configure server, passing the necessary configs
	e := NewServer(cfg)

	// Initialize routes
	InitializeRoutes(e, queries)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.WithContext(context.Background(), "startup").Info(fmt.Sprintf("Server starting on %s", serverAddr), nil)
	// Start the server in a goroutine so it doesn't block here.
	// The actual listening and error handling will be done in main.go.
	go func() {
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			logger.WithContext(context.Background(), "startup").Fatal("Server failed to start", err, nil)
		}
	}()
	return e, db, nil
}

// NewServer creates a new Echo server instance with middleware, using loaded config
func NewServer(cfg *config.AppConfig) *echo.Echo {
	e := echo.New()

	// Set custom error handler
	e.HTTPErrorHandler = errors.ErrorHandler

	// Metrics Middleware - should be early in the chain
	e.Use(metrics.Middleware)

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
	e.Use(middleware.BodyLimit(cfg.BodyLimit))

	// CORS Middleware with secure configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORSOptions.AllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate Limiter with more robust configuration
	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(cfg.RateLimitOptions.Requests),
			Burst:     cfg.RateLimitOptions.Burst,
			ExpiresIn: time.Minute * 1,
		},
	)
	e.Use(middleware.RateLimiter(store))

	return e
}

// InitializeRoutes sets up all the routes for the server
func InitializeRoutes(e *echo.Echo, queries *database.Queries) {
	// Add the routes here
	routes.HealthCheckRoutes(e, context.Background(), queries) // Use context.Background()
	routes.RegisterTodoRoutes(e, queries)

	// Add metrics endpoint, using the custom AppRegistry from the metrics package
	// TODO: When doing the implementation, add security priority for metrics endpoint so that it is not exposed to the public
	e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(metrics.AppRegistry, promhttp.HandlerOpts{})))
}
