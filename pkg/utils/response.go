package utils

import (
	"net/http"

	"arquivolivre.com.br/otel/internal/models"

	"github.com/gin-gonic/gin"
)

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

func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func SendBadRequest(c *gin.Context, message string) {
	SendError(c, http.StatusBadRequest, message)
}

func SendNotFound(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, message)
}

func SendInternalError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, message)
}

func SendConflict(c *gin.Context, message string) {
	SendError(c, http.StatusConflict, message)
}
