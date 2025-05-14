package server

import (
	"context"
	"go-echo-server-template/internal/config"
	appErrors "go-echo-server-template/internal/errors"
	"go-echo-server-template/internal/logger"
	"go-echo-server-template/routes"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/joho/godotenv"
)

func InitServer() {
	// Create a background context for server initialization
	ctx := context.Background()

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	// Initialize logger first
	if err := logger.Initialize(os.Getenv("APP_ENV")); err != nil {
		log.Fatal("Failed to initialize logger: ", err)
	}

	// Create a logger instance for server initialization
	log := logger.WithContext(ctx, "server_init")

	port := os.Getenv("PORT")
	if port == "" {
		log.Error("PORT environment variable not found", nil, nil)
		os.Exit(1)
	}

	// Initialize database configuration
	dbConfig, err := config.NewDatabaseConfig()
	if err != nil {
		log.Error("Failed to create database configuration", err, nil)
		os.Exit(1)
	}

	// Initialize database connection
	db, err := config.InitializeDatabase(dbConfig)
	if err != nil {
		log.Error("Failed to initialize database", err, nil)
		os.Exit(1)
	}

	log.Debug("Database connection established", map[string]interface{}{
		"max_open_conns":    dbConfig.MaxOpenConns,
		"max_idle_conns":    dbConfig.MaxIdleConns,
		"conn_max_lifetime": dbConfig.ConnMaxLifetime,
	})

	// Echo webframework
	e := echo.New()

	// Set custom error handler
	e.HTTPErrorHandler = appErrors.HTTPErrorHandler()

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
	e.Use(middleware.BodyLimit(os.Getenv("BODY_LIMIT")))

	// CORS Middleware with secure configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{os.Getenv("ALLOWED_ORIGINS")}, // Configure this in .env
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Rate Limiter with more robust configuration
	rateLimit, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS"))
	burstLimit, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_BURST"))

	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(rateLimit),
			Burst:     burstLimit,
			ExpiresIn: time.Minute * 1,
		},
	)

	e.Use(middleware.RateLimiter(store))

	// Routes
	routes.HealthCheckRoutes(e, ctx, db)
	routes.InitTodoRouter(e, ctx, db)

	log.Info("Server starting", map[string]interface{}{
		"port": port,
	})

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}
