package handlers

import (
	"net/http"
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/models"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/services"
	"github.com/gin-gonic/gin"
)

// StatusStore is the subset of status persistence behavior needed by the handler.
type StatusStore interface {
	GetStatus(requestID string) (string, error)
	SetStatus(requestID, status string) error
}

// NotificationHandler handles notification-related requests.
type NotificationHandler struct {
	idempotencyService *services.IdempotencyService
	userClient         *services.UserClient
	publisher          *services.Publisher
	statusStore        StatusStore
	templateClient     *services.TemplateClient
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(
	idempotencyService *services.IdempotencyService,
	userClient *services.UserClient,
	publisher *services.Publisher,
	statusStore StatusStore,
	templateClient *services.TemplateClient,
) *NotificationHandler {
	return &NotificationHandler{
		idempotencyService: idempotencyService,
		userClient:         userClient,
		publisher:          publisher,
		statusStore:        statusStore,
		templateClient:     templateClient,
	}
}

// SendNotification handles the request to send a notification.
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req models.SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check for idempotency
	if isDuplicate, err := h.idempotencyService.IsDuplicate(req.RequestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check idempotency"})
		return
	} else if isDuplicate {
		// If it's a duplicate, return the previous status
		status, err := h.statusStore.GetStatus(req.RequestID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status for duplicate request"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": status})
		return
	}

	// Get user preferences
	prefs, err := h.userClient.GetPreferences(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user preferences"})
		return
	}

	// Check if the user has opted out of this notification channel
	if (req.Channel == "email" && !prefs.AllowEmail) || (req.Channel == "push" && !prefs.AllowPush) {
		c.JSON(http.StatusOK, gin.H{"status": "skipped", "message": "user has opted out of this channel"})
		return
	}

	// Get template
	tpl, err := h.templateClient.GetTemplate(req.TemplateSlug, prefs.Locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get template"})
		return
	}

	// Create message envelope
	envelope := models.MessageEnvelope{
		RequestID:     req.RequestID,
		CorrelationID: c.GetString("correlation_id"),
		CreatedAt:     time.Now(),
		Channel:       req.Channel,
		User: models.User{
			ID:         req.UserID,
			Email:      prefs.Email,
			Locale:     prefs.Locale,
			PushTokens: prefs.PushTokens,
		},
		Template: models.Template{
			Slug:    req.TemplateSlug,
			Locale:  prefs.Locale,
			Version: tpl.Version,
		},
		Variables:         req.Variables,
		ProviderOverrides: req.Metadata,
		RetryCount:        0,
	}

	// Publish message to RabbitMQ
	if err := h.publisher.Publish(&envelope); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish message"})
		return
	}

	// Store initial status
	if err := h.statusStore.SetStatus(req.RequestID, "queued"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set initial status"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}
