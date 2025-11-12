package models

import "errors"

// SendRequest represents the request to send a notification.
type SendRequest struct {
	RequestID        string                 `json:"request_id" binding:"required,uuid4"`
	UserID           string                 `json:"user_id" binding:"required,uuid4"`
	Channel          string                 `json:"channel" binding:"omitempty,oneof=email push"`
	NotificationType string                 `json:"notification_type" binding:"omitempty,oneof=email push"`
	TemplateSlug     string                 `json:"template_slug" binding:"omitempty"`
	TemplateCode     string                 `json:"template_code" binding:"omitempty"`
	Variables        map[string]interface{} `json:"variables"`
	Priority         string                 `json:"priority"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// Normalize harmonizes legacy aliases and defaults.
func (r *SendRequest) Normalize() error {
	if r.Channel == "" {
		r.Channel = r.NotificationType
	}
	if r.Channel == "" {
		return errors.New("channel or notification_type is required")
	}

	if r.TemplateSlug == "" {
		r.TemplateSlug = r.TemplateCode
	}
	if r.TemplateSlug == "" {
		return errors.New("template_slug or template_code is required")
	}

	if r.Priority == "" {
		r.Priority = "normal"
	}
	if r.Variables == nil {
		r.Variables = map[string]interface{}{}
	}
	if r.Metadata == nil {
		r.Metadata = map[string]interface{}{}
	}
	return nil
}
