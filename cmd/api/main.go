package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example/otel/internal/config"
	"example/otel/internal/database"
	"example/otel/internal/handlers"
	"example/otel/internal/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize structured logging first
	logging.InitGlobalLogger()
	logger := logging.GetLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize OpenTelemetry
	telemetryCfg := config.GetTelemetryConfig()
	telemetryProvider, err := config.InitTelemetry(telemetryCfg)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := telemetryProvider.Shutdown(shutdownCtx); err != nil {
			logger.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("Error shutting down telemetry")
		}
	}()

	logger.WithFields(map[string]interface{}{
		"service_name":            telemetryCfg.ServiceName,
		"service_version":         telemetryCfg.ServiceVersion,
		"tracing_enabled":         telemetryCfg.EnableTracing,
		"metrics_enabled":         telemetryCfg.EnableMetrics,
		"logging_enabled":         telemetryCfg.EnableLogging,
		"runtime_metrics_enabled": telemetryCfg.EnableRuntimeMetrics,
	}).Info("OpenTelemetry initialized successfully")

	// Setup OpenTelemetry logging hook if logging is enabled
	logger.WithFields(map[string]interface{}{
		"enable_logging":      telemetryCfg.EnableLogging,
		"logger_provider_nil": telemetryProvider.LoggerProvider == nil,
	}).Info("Checking logging configuration")

	if telemetryCfg.EnableLogging && telemetryProvider.LoggerProvider != nil {
		logging.SetupOtelHook(telemetryProvider.LoggerProvider)
		logger.Info("OpenTelemetry logging hook configured")
	} else {
		logger.WithFields(map[string]interface{}{
			"enable_logging":      telemetryCfg.EnableLogging,
			"logger_provider_nil": telemetryProvider.LoggerProvider == nil,
		}).Warn("OpenTelemetry logging hook not configured")
	}

	// Set Gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Start database connection monitoring
	monitorCtx, cancelMonitor := context.WithCancel(context.Background())
	defer cancelMonitor()
	db.StartConnectionMonitoring(monitorCtx, 30*time.Second)

	// Setup routes
	router := handlers.SetupRoutes(db)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	// if err := server.Shutdown(ctx); err != nil {
	// 	log.Fatalf("Server forced to shutdown: %v", err)
	// }

	log.Println("Server exited")
}
