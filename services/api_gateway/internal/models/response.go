package models

// PaginationMeta captures pagination metadata for list responses.
type PaginationMeta struct {
	Total       int  `json:"total"`
	Limit       int  `json:"limit"`
	Page        int  `json:"page"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// ResponseEnvelope is the canonical response shape for the API gateway.
type ResponseEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    interface{}     `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

// UserPreferences represents the user's notification preferences and delivery data.
type UserPreferences struct {
	AllowEmail bool        `json:"allow_email"`
	AllowPush  bool        `json:"allow_push"`
	Email      string      `json:"email"`
	Locale     string      `json:"locale"`
	PushTokens []PushToken `json:"push_tokens"`
}
