package models

// SendRequest represents the request to send a notification.
type SendRequest struct {
	RequestID    string                 `json:"request_id" binding:"required,uuid4"`
	UserID       string                 `json:"user_id" binding:"required,uuid4"`
	Channel      string                 `json:"channel" binding:"required,oneof=email push"`
	TemplateSlug string                 `json:"template_slug" binding:"required"`
	Variables    map[string]interface{} `json:"variables"`
	Priority     string                 `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata"`
}
