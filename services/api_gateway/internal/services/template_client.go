package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/models"
)

// TemplateClient is a client for the template service.
type TemplateClient struct {
	baseURL string
	client  *http.Client
}

// NewTemplateClient creates a new TemplateClient.
func NewTemplateClient(baseURL string) *TemplateClient {
	return &TemplateClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// GetTemplate retrieves a template from the template service.
func (c *TemplateClient) GetTemplate(slug, locale string) (*models.Template, error) {
	url := fmt.Sprintf("%s/v1/templates/%s/active?locale=%s", c.baseURL, slug, locale)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("template service returned non-200 status code: %d", resp.StatusCode)
	}

	var tpl models.Template
	if err := json.NewDecoder(resp.Body).Decode(&tpl); err != nil {
		return nil, err
	}

	return &tpl, nil
}
