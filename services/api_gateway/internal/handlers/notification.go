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
	SetStatusWithProvider(requestID, status, provider, detail string) error
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
		respondValidationError(c, err)
		return
	}
	if err := req.Normalize(); err != nil {
		respondValidationError(c, err)
		return
	}

	// Check for idempotency
	if isDuplicate, err := h.idempotencyService.IsDuplicate(req.RequestID); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to check idempotency", err)
		return
	} else if isDuplicate {
		// If it's a duplicate, return the previous status
		status, err := h.statusStore.GetStatus(req.RequestID)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "failed to get status for duplicate request", err)
			return
		}
		respondSuccess(c, http.StatusOK, "duplicate request", gin.H{
			"request_id": req.RequestID,
			"status":     status,
		})
		return
	}

	// Get user preferences
	prefs, err := h.userClient.GetPreferences(req.UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get user preferences", err)
		return
	}

	// Check if the user has opted out of this notification channel
	if (req.Channel == "email" && !prefs.AllowEmail) || (req.Channel == "push" && !prefs.AllowPush) {
		_ = h.statusStore.SetStatusWithProvider(req.RequestID, "skipped", "", "user opted out")
		respondSuccess(c, http.StatusOK, "user has opted out of this channel", gin.H{
			"request_id": req.RequestID,
			"status":     "skipped",
		})
		return
	}

	// Get template
	tpl, err := h.templateClient.GetTemplate(req.TemplateSlug, prefs.Locale)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get template", err)
		return
	}

	template := *tpl
	template.Slug = req.TemplateSlug
	template.Locale = prefs.Locale

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
		Template:          template,
		Variables:         req.Variables,
		ProviderOverrides: req.Metadata,
		RetryCount:        0,
	}

	// Publish message to RabbitMQ
	if err := h.publisher.Publish(&envelope); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to publish message", err)
		return
	}

	// Store initial status
	if err := h.statusStore.SetStatus(req.RequestID, "queued"); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to set initial status", err)
		return
	}

	respondSuccess(c, http.StatusAccepted, "notification queued", gin.H{
		"request_id": req.RequestID,
		"status":     "queued",
	})
}
