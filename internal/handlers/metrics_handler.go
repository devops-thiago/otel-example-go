package handlers

import (
	"net/http"

	"example/otel/internal/database"

	"github.com/gin-gonic/gin"
)

// MetricsHandler handles metrics-related requests
type MetricsHandler struct {
	db *database.DB
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(db *database.DB) *MetricsHandler {
	return &MetricsHandler{db: db}
}

// GetMetrics handles GET /metrics - returns database and application metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	// Get database health status
	healthErr := h.db.Health()

	// Get detailed database statistics
	dbStats := h.db.GetDetailedStats()

	// Prepare response
	response := gin.H{
		"database": gin.H{
			"healthy": healthErr == nil,
			"error": func() string {
				if healthErr != nil {
					return healthErr.Error()
				}
				return ""
			}(),
			"stats": dbStats,
		},
		"application": gin.H{
			"status": "running",
		},
		"message": "Application and database metrics",
	}

	statusCode := http.StatusOK
	if healthErr != nil {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
