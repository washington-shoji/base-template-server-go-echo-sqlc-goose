package server

import (
	"context"
	"database/sql"
	"go-echo-server-template/internal/database"
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
	_ "github.com/lib/pq"
)

func InitServer() {
	godotenv.Load(".env")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT could not be found in the process enviroment")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL could not be found in the process enviroment")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cannot connect to the database: ", err)
	}

	db := database.New(conn)

	// Echo webframework
	e := echo.New()

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

	// Logger Middleware with custom format
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}))

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

	ctx := context.Background()

	// Routes
	routes.HealthCheckRoutes(e, ctx, db)
	routes.InitTodoRouter(e, ctx, db)

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
	log.Printf("Server starting on port %v", port)
}
