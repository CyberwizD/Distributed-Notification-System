package handlers

import (
	"net/http"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/repository"
	"github.com/gin-gonic/gin"
)

// StatusHandler handles status-related requests.
type StatusHandler struct {
	statusStore *repository.StatusStore
}

// NewStatusHandler creates a new StatusHandler.
func NewStatusHandler(statusStore *repository.StatusStore) *StatusHandler {
	return &StatusHandler{statusStore: statusStore}
}

// GetStatus handles the request to get the status of a notification.
func (h *StatusHandler) GetStatus(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id is required"})
		return
	}

	status, err := h.statusStore.GetStatus(requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"request_id": requestID, "status": status})
}
