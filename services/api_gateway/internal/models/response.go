package models

// UserPreferences represents the user's notification preferences and delivery data.
type UserPreferences struct {
	AllowEmail bool        `json:"allow_email"`
	AllowPush  bool        `json:"allow_push"`
	Email      string      `json:"email"`
	Locale     string      `json:"locale"`
	PushTokens []PushToken `json:"push_tokens"`
}
