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

// UserClient is a client for the user service with Redis-backed caching.
type UserClient struct {
	baseURL  string
	client   *http.Client
	cache    *repository.RedisRepository
	cacheTTL time.Duration
}

// NewUserClient creates a new UserClient.
func NewUserClient(baseURL string, cache *repository.RedisRepository, cacheTTL time.Duration) *UserClient {
	return &UserClient{
		baseURL: baseURL,
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

	url := fmt.Sprintf("%s/v1/users/%s/preferences", c.baseURL, userID)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned non-200 status code: %d", resp.StatusCode)
	}

	var prefs models.UserPreferences
	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return nil, err
	}

	if c.cache != nil && c.cacheTTL > 0 {
		_ = c.cache.SetJSON(ctx, cacheKey, &prefs, c.cacheTTL)
	}

	return &prefs, nil
}
