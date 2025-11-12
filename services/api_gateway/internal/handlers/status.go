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
		respondError(c, http.StatusBadRequest, "request_id is required", nil)
		return
	}

	status, err := h.statusStore.GetStatus(requestID)
	if err != nil {
		respondError(c, http.StatusNotFound, "notification not found", err)
		return
	}

	respondSuccess(c, http.StatusOK, "notification status retrieved", gin.H{
		"request_id": requestID,
		"status":     status,
	})
}
