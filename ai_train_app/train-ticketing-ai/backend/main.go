package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"train-ticketing-ai/config"
	"train-ticketing-ai/database"
	"train-ticketing-ai/handlers"
	"train-ticketing-ai/services"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting Train Ticketing System")
	log.Printf("AI Provider: %s", cfg.AIProvider)

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations check
	if err := database.RunMigrations(database.GetDB()); err != nil {
		log.Printf("Migration check warning: %v", err)
	}

	// Initialize AI service
	services.InitAIService(cfg)

	// Setup Gin router
	router := setupRouter()

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter() *gin.Engine {
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Station routes
		api.GET("/stations", handlers.GetStations)

		// Train routes
		api.POST("/search", handlers.SearchTrains)
		api.GET("/trains/:id", handlers.GetTrain)

		// Booking routes
		api.POST("/bookings", handlers.CreateBooking)
		api.GET("/bookings/:ref", handlers.GetBooking)
		api.DELETE("/bookings/:ref", handlers.CancelBooking)

		// AI chat routes
		api.POST("/ai/chat", handlers.ChatWithAI)
	}

	// Serve static files (frontend)
	router.Static("/css", "./frontend/css")
	router.Static("/js", "./frontend/js")
	router.StaticFile("/", "./frontend/index.html")
	router.StaticFile("/index.html", "./frontend/index.html")

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
	})

	return router
}
