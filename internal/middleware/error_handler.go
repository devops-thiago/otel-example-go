package middleware

import (
	"net/http"

	"example/otel/internal/models"

	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware to handle errors consistently
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Success: false,
					Error:   "Invalid request data: " + err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Success: false,
					Error:   err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Success: false,
					Error:   "Internal server error",
				})
			}
		}
	}
}
