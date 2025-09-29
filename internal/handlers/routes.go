package handlers

import (
	"arquivolivre.com.br/otel/internal/database"
	"arquivolivre.com.br/otel/internal/logging"
	"arquivolivre.com.br/otel/internal/middleware"
	"arquivolivre.com.br/otel/internal/repository"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(db *database.DB) *gin.Engine {
	router := gin.New()

	telemetryMiddleware := middleware.NewTelemetryMiddleware("otel-example-api")

	logger := logging.NewLogger()

	router.Use(logger.Middleware())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(telemetryMiddleware.GinMiddleware())
	router.Use(telemetryMiddleware.MetricsMiddleware())
	router.Use(middleware.ErrorHandler())

	userRepo := repository.NewUserRepository(db)

	healthHandler := NewHealthHandler(db)
	userHandler := NewUserHandler(userRepo)
	metricsHandler := NewMetricsHandler(db)

	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)

	router.GET("/metrics", metricsHandler.GetMetrics)

	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "OpenTelemetry Example API",
				"version": "1.0.0",
				"status":  "running",
			})
		})

		users := api.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	return router
}
