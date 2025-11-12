package models

import "time"

// MessageEnvelope is the payload produced by the API gateway and consumed by the push service.
type MessageEnvelope struct {
	RequestID         string                 `json:"request_id"`
	CorrelationID     string                 `json:"correlation_id"`
	CreatedAt         time.Time              `json:"created_at"`
	Channel           string                 `json:"channel"`
	User              User                   `json:"user"`
	Template          Template               `json:"template"`
	Variables         map[string]interface{} `json:"variables"`
	ProviderOverrides map[string]interface{} `json:"provider_overrides,omitempty"`
	RetryCount        int                    `json:"retry_count"`
}

type User struct {
	ID         string      `json:"id"`
	Email      string      `json:"email"`
	Locale     string      `json:"locale"`
	PushTokens []PushToken `json:"push_tokens"`
}

type PushToken struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
	Provider string `json:"provider,omitempty"`
}

type Template struct {
	Slug    string `json:"slug"`
	Locale  string `json:"locale"`
	Version int    `json:"version"`
	Title   string `json:"title,omitempty"`
	Body    string `json:"body,omitempty"`
}

// RenderedTemplate is the final text after template substitution.
type RenderedTemplate struct {
	Title string
	Body  string
}
