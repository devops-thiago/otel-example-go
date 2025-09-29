package handlers

import (
    "net/http"

    "example/otel/internal/models"

    "github.com/gin-gonic/gin"
)

// DBHealth defines the minimal contract used by the health handler
type DBHealth interface {
    Health() error
}

// HealthHandler handles health check requests
type HealthHandler struct {
    db DBHealth
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db DBHealth) *HealthHandler {
    return &HealthHandler{db: db}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Check database health
	if err := h.db.Health(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Service is healthy",
		Data: map[string]string{
			"status":   "healthy",
			"database": "connected",
		},
	})
}

// ReadinessCheck handles GET /ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Perform more comprehensive checks here
	// For now, just check database
	if err := h.db.Health(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Success: false,
			Error:   "Service not ready",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Service is ready",
	})
}
