package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck handles the health check endpoint.
func HealthCheck(c *gin.Context) {
	respondSuccess(c, http.StatusOK, "api-gateway healthy", gin.H{
		"status": "ok",
	})
}
