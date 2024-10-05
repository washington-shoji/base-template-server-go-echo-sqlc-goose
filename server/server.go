package server

import (
	"context"
	"database/sql"
	"go-echo-server-template/internal/database"
	"go-echo-server-template/routes"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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

	// Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://*", "http://*"}, // MAke sure to updated the allowed domains in production
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(30)))

	ctx := context.Background()

	// Routes
	routes.HealthCheckRoutes(e, ctx, db)
	routes.InitTodoRouter(e, ctx, db)

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
	log.Printf("Server starting on port %v", port)
}
