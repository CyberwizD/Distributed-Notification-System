package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/models"
)

// TemplateClient fetches templates from the template service.
type TemplateClient struct {
	baseURL string
	client  *http.Client
}

func NewTemplateClient(baseURL string, timeout time.Duration) *TemplateClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &TemplateClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *TemplateClient) Fetch(ctx context.Context, slug, locale string) (*models.Template, error) {
	if locale == "" {
		locale = "en"
	}
	path := fmt.Sprintf("%s/v1/templates/%s/active?locale=%s",
		c.baseURL,
		url.PathEscape(slug),
		url.QueryEscape(locale),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("template service returned %d", resp.StatusCode)
	}

	var tpl models.Template
	if err := json.NewDecoder(resp.Body).Decode(&tpl); err != nil {
		return nil, err
	}
	return &tpl, nil
}
