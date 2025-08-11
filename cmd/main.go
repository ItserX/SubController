package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/ItserX/rest/docs"
	"github.com/ItserX/rest/internal/handlers"
	"github.com/ItserX/rest/internal/logger"
	"github.com/ItserX/rest/internal/storage"
)

// @title API сервиса подписок
// @version 1.0
// @description Сервис для управления подписками пользователей
// @host localhost:8080
// @BasePath /api
// @schemes http

func connectDB() *sql.DB {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	logger.Logger.Infow("Connecting to database",
		"host", os.Getenv("DB_HOST"),
		"port", os.Getenv("DB_PORT"),
		"dbname", os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Logger.Fatalw("Database connection failed", "error", err)
	}
	err = db.Ping()
	if err != nil {
		logger.Logger.Fatalw("Database ping failed", "error", err)
	}

	logger.Logger.Info("Successfully connected to database")
	return db
}

func startServer(db *sql.DB) {
	gin.SetMode(os.Getenv("GIN_MODE"))

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(loggingMiddleware())

	h := handlers.Handler{
		Repo: storage.NewPostgresRepository(db),
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.GET("/:id", h.GetSub)
			subscriptions.POST("", h.CreateSub)
			subscriptions.PUT("/:id", h.UpdateSub)
			subscriptions.DELETE("/:id", h.DeleteSub)
			subscriptions.GET("/list", h.ListSubs)
			subscriptions.GET("/totalCost", h.GetTotalCost)
		}
	}

	port := ":" + os.Getenv("SERVER_PORT")
	logger.Logger.Infow("Starting server", "port", port)
	err := r.Run(port)
	if err != nil {
		logger.Logger.Fatalw("Server failed to start", "error", err)
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Logger.Infow("Incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"ip", c.ClientIP(),
		)
		c.Next()
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = logger.SetupLogger()
	if err != nil {
		log.Fatal("Failed to initialize logger: ", err)
	}
	defer logger.Logger.Sync()

	db := connectDB()
	defer db.Close()

	startServer(db)
}
