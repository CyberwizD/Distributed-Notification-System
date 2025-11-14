package models

import "time"

// MessageEnvelope represents the common structure for messages sent via RabbitMQ.
type MessageEnvelope struct {
	RequestID         string                 `json:"request_id"`
	CorrelationID     string                 `json:"correlation_id"`
	CreatedAt         time.Time              `json:"created_at"`
	Channel           string                 `json:"channel"` // "email" or "push"
	User              User                   `json:"user"`
	Template          Template               `json:"template"`
	Variables         map[string]interface{} `json:"variables"`
	ProviderOverrides map[string]interface{} `json:"provider_overrides,omitempty"`
	RetryCount        int                    `json:"retry_count"`
}

// User represents user information in the message.
type User struct {
	ID         string      `json:"id"`
	Email      string      `json:"email"`
	Locale     string      `json:"locale"`
	PushTokens []PushToken `json:"push_tokens"`
}

// PushToken represents a user's device token.
type PushToken struct {
	Token    string `json:"token"`
	Platform string `json:"platform"` // "android", "ios", "web"
}

// Template represents template information.
type Template struct {
	Slug    string `json:"slug"`
	Locale  string `json:"locale"`
	Version int    `json:"version"`
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}
