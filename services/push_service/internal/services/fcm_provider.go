package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/models"
)

// FCMProvider sends notifications via Firebase Cloud Messaging.
type FCMProvider struct {
	serverKey string
	endpoint  string
	client    *http.Client
	logger    *slog.Logger
	timeout   time.Duration
}

func NewFCMProvider(serverKey, endpoint string, timeout time.Duration, logger *slog.Logger) *FCMProvider {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &FCMProvider{
		serverKey: serverKey,
		endpoint:  endpoint,
		client: &http.Client{
			Timeout: timeout,
		},
		logger:  logger,
		timeout: timeout,
	}
}

func (p *FCMProvider) Name() string {
	return "fcm"
}

func (p *FCMProvider) Send(ctx context.Context, payload *PushPayload) ([]models.PushResult, error) {
	if len(payload.Tokens) == 0 {
		return nil, fmt.Errorf("fcm: no tokens supplied")
	}

	regIDs := make([]string, 0, len(payload.Tokens))
	for _, token := range payload.Tokens {
		if token.Token == "" {
			continue
		}
		regIDs = append(regIDs, token.Token)
	}

	if len(regIDs) == 0 {
		return nil, fmt.Errorf("fcm: tokens were empty")
	}

	reqMap := map[string]interface{}{
		"registration_ids": regIDs,
		"notification": map[string]string{
			"title": payload.Title,
			"body":  payload.Body,
		},
	}
	if len(payload.Data) > 0 {
		reqMap["data"] = payload.Data
	}
	if overrides := providerOverrides(payload.Overrides, "fcm"); overrides != nil {
		mergeMaps(reqMap, overrides)
	}

	body, err := json.Marshal(reqMap)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+p.serverKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fcm: received status %d", resp.StatusCode)
	}

	var fcmResp fcmResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
		return nil, err
	}

	results := make([]models.PushResult, 0, len(fcmResp.Results))
	for idx, res := range fcmResp.Results {
		token := ""
		if idx < len(regIDs) {
			token = regIDs[idx]
		}
		status := models.ResultDelivered
		if res.Error != "" {
			status = models.ResultFailed
		}

		results = append(results, models.PushResult{
			Token:     token,
			Provider:  p.Name(),
			Status:    status,
			MessageID: res.MessageID,
			Error:     res.Error,
		})
	}

	return results, nil
}

type fcmResponse struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
	Results []struct {
		MessageID string `json:"message_id"`
		Error     string `json:"error"`
	} `json:"results"`
}

func providerOverrides(overrides map[string]interface{}, key string) map[string]interface{} {
	if overrides == nil {
		return nil
	}
	if raw, ok := overrides[key]; ok {
		if cast, ok := raw.(map[string]interface{}); ok {
			return cast
		}
	}
	return nil
}

func mergeMaps(dst map[string]interface{}, src map[string]interface{}) {
	for key, value := range src {
		if nestedSrc, ok := value.(map[string]interface{}); ok {
			if nestedDst, ok := dst[key].(map[string]interface{}); ok {
				mergeMaps(nestedDst, nestedSrc)
				continue
			}
		}
		dst[key] = value
	}
}
