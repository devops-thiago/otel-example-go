package handlers

import (
	"example/otel/internal/database"
	"example/otel/internal/logging"
	"example/otel/internal/middleware"
	"example/otel/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes with OpenTelemetry instrumentation
func SetupRoutes(db *database.DB) *gin.Engine {
	// Create Gin router
	router := gin.New()

	// Initialize telemetry middleware
	telemetryMiddleware := middleware.NewTelemetryMiddleware("otel-example-api")

	// Initialize structured logging
	logger := logging.NewLogger()

	// Add middleware (order matters)
	router.Use(logger.Middleware()) // Structured logging
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(telemetryMiddleware.GinMiddleware())     // OpenTelemetry tracing
	router.Use(telemetryMiddleware.MetricsMiddleware()) // Custom metrics
	router.Use(middleware.ErrorHandler())

	// Initialize repositories
    userRepo := repository.NewUserRepository(db)

	// Initialize handlers
    healthHandler := NewHealthHandler(db)
    userHandler := NewUserHandler(userRepo)
	metricsHandler := NewMetricsHandler(db)

	// Health check routes
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)

	// Metrics endpoint for Prometheus scraping
	router.GET("/metrics", metricsHandler.GetMetrics)

	// API routes
	api := router.Group("/api")
	{
		// API info
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "OpenTelemetry Example API",
				"version": "1.0.0",
				"status":  "running",
			})
		})

		// User routes
		users := api.Group("/users")
		{
			users.GET("", userHandler.GetUsers)          // GET /api/users
			users.POST("", userHandler.CreateUser)       // POST /api/users
			users.GET("/:id", userHandler.GetUser)       // GET /api/users/:id
			users.PUT("/:id", userHandler.UpdateUser)    // PUT /api/users/:id
			users.DELETE("/:id", userHandler.DeleteUser) // DELETE /api/users/:id
		}
	}

	return router
}
