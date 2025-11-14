package services

import (
	"context"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/models"
)

// PushPayload is the fully rendered payload handed to a provider.
type PushPayload struct {
	Tokens    []models.PushToken
	Title     string
	Body      string
	Data      map[string]string
	Overrides map[string]interface{}
}

// PushProvider represents a downstream push provider (FCM, OneSignal, etc).
type PushProvider interface {
	Name() string
	Send(ctx context.Context, payload *PushPayload) ([]models.PushResult, error)
}
