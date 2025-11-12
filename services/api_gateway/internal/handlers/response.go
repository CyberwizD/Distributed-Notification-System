package handlers

import (
	"net/http"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/models"
	"github.com/gin-gonic/gin"
)

func respondSuccess(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, models.ResponseEnvelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func respondError(c *gin.Context, status int, message string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	c.JSON(status, models.ResponseEnvelope{
		Success: false,
		Message: message,
		Error:   errMsg,
	})
}

func respondValidationError(c *gin.Context, err error) {
	respondError(c, http.StatusBadRequest, "validation failed", err)
}
