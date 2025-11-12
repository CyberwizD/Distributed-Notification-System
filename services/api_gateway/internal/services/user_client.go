package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/models"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/repository"
)

type userClientResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    *userProfileDTO `json:"data"`
	Error   string          `json:"error"`
}

type userProfileDTO struct {
	UserID       string           `json:"userId"`
	Email        string           `json:"email"`
	Preferences  preferenceDTO    `json:"preferences"`
	DeviceTokens []deviceTokenDTO `json:"deviceTokens"`
}

type preferenceDTO struct {
	EmailEnabled bool   `json:"emailEnabled"`
	PushEnabled  bool   `json:"pushEnabled"`
	Language     string `json:"language"`
}

type deviceTokenDTO struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

// UserClient is a client for the user service with Redis-backed caching.
type UserClient struct {
	baseURL        string
	internalAPIKey string
	client         *http.Client
	cache          *repository.RedisRepository
	cacheTTL       time.Duration
}

// NewUserClient creates a new UserClient.
func NewUserClient(baseURL, internalAPIKey string, cache *repository.RedisRepository, cacheTTL time.Duration) *UserClient {
	return &UserClient{
		baseURL:        baseURL,
		internalAPIKey: internalAPIKey,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache:    cache,
		cacheTTL: cacheTTL,
	}
}

// GetPreferences retrieves user preferences from the user service, falling back to cache when possible.
func (c *UserClient) GetPreferences(userID string) (*models.UserPreferences, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:prefs:%s", userID)

	if c.cache != nil && c.cacheTTL > 0 {
		var cached models.UserPreferences
		if ok, err := c.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			return &cached, nil
		}
	}

	url := fmt.Sprintf("%s/internal/users/%s/notification-profile", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-API-Key", c.internalAPIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status code: %d", resp.StatusCode)
	}

	var envelope userClientResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	if !envelope.Success || envelope.Data == nil {
		return nil, fmt.Errorf("user service error: %s", envelope.Message)
	}

	prefs := &models.UserPreferences{
		AllowEmail: envelope.Data.Preferences.EmailEnabled,
		AllowPush:  envelope.Data.Preferences.PushEnabled,
		Email:      envelope.Data.Email,
		Locale:     envelope.Data.Preferences.Language,
	}
	for _, token := range envelope.Data.DeviceTokens {
		if token.Token == "" {
			continue
		}
		prefs.PushTokens = append(prefs.PushTokens, models.PushToken{
			Token:    token.Token,
			Platform: token.Platform,
		})
	}

	if c.cache != nil && c.cacheTTL > 0 {
		_ = c.cache.SetJSON(ctx, cacheKey, prefs, c.cacheTTL)
	}

	return prefs, nil
}
