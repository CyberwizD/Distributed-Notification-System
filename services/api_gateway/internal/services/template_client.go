package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/models"
)

type templateResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    *templateDTO `json:"data"`
	Error   string       `json:"error"`
}

type templateDTO struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

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
	if locale == "" {
		locale = "en"
	}
	endpoint := fmt.Sprintf("%s/v1/templates/%s/active?locale=%s",
		c.baseURL,
		url.PathEscape(slug),
		url.QueryEscape(locale),
	)
	resp, err := c.client.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("template service returned status: %d", resp.StatusCode)
	}

	var envelope templateResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	if !envelope.Success || envelope.Data == nil {
		return nil, fmt.Errorf("template service error: %s", envelope.Message)
	}

	return &models.Template{
		Slug:    slug,
		Locale:  locale,
		Version: 0,
		Subject: envelope.Data.Subject,
		Body:    envelope.Data.Body,
	}, nil
}
