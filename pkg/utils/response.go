package utils

import (
	"net/http"

	"arquivolivre.com.br/otel/internal/models"

	"github.com/gin-gonic/gin"
)

// SendSuccess sends a successful response
func SendSuccess(c *gin.Context, data interface{}, message ...string) {
	response := models.SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(http.StatusOK, response)
}

// SendCreated sends a created response
func SendCreated(c *gin.Context, data interface{}, message ...string) {
	response := models.SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(http.StatusCreated, response)
}

// SendError sends an error response
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

// SendBadRequest sends a bad request error
func SendBadRequest(c *gin.Context, message string) {
	SendError(c, http.StatusBadRequest, message)
}

// SendNotFound sends a not found error
func SendNotFound(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, message)
}

// SendInternalError sends an internal server error
func SendInternalError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, message)
}

// SendConflict sends a conflict error
func SendConflict(c *gin.Context, message string) {
	SendError(c, http.StatusConflict, message)
}
