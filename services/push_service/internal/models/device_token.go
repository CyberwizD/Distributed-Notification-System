package models

import "strings"

// DeviceToken represents a user device that can receive push notifications.
type DeviceToken struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
	Provider string `json:"provider,omitempty"`
}

// PlatformCategory normalizes a platform string to one of the supported categories.
func PlatformCategory(platform string) string {
	switch strings.ToLower(platform) {
	case "android", "ios":
		return "mobile"
	case "web":
		return "web"
	default:
		return "unknown"
	}
}
