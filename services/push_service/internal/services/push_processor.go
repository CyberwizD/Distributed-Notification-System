package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/models"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/repository"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/pkg/metrics"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/pkg/retry"
)

type PushProcessor struct {
	templateClient *TemplateClient
	fcm            PushProvider
	statusUpdater  *StatusUpdater
	cache          *repository.RedisRepository
	metrics        *metrics.Metrics
	logger         *slog.Logger
	retryCfg       retry.Config
}

func NewPushProcessor(
	templateClient *TemplateClient,
	fcm PushProvider,
	statusUpdater *StatusUpdater,
	cache *repository.RedisRepository,
	metrics *metrics.Metrics,
	logger *slog.Logger,
	retryCfg retry.Config,
) *PushProcessor {
	return &PushProcessor{
		templateClient: templateClient,
		fcm:            fcm,
		statusUpdater:  statusUpdater,
		cache:          cache,
		metrics:        metrics,
		logger:         logger,
		retryCfg:       retryCfg,
	}
}

func (p *PushProcessor) Process(ctx context.Context, envelope *models.MessageEnvelope) error {
	if envelope.Channel != "push" {
		return fmt.Errorf("unexpected channel %s", envelope.Channel)
	}
	p.metrics.IncConsumed()

	activeTokens, err := p.filterTokens(ctx, envelope.User.PushTokens)
	if err != nil {
		p.logger.Error("failed to filter tokens", slog.Any("error", err))
		return err
	}
	if len(activeTokens) == 0 {
		err := fmt.Errorf("no valid push tokens")
		p.statusUpdater.MarkFailed(ctx, envelope.RequestID, p.fcm.Name(), err.Error())
		p.metrics.IncFailed()
		return err
	}

	tpl, err := p.templateClient.Fetch(ctx, envelope.Template.Slug, localeFromEnvelope(envelope))
	if err != nil {
		p.statusUpdater.MarkFailed(ctx, envelope.RequestID, p.fcm.Name(), err.Error())
		p.metrics.IncFailed()
		return err
	}

	titleTemplate := tpl.Subject
	if titleTemplate == "" {
		titleTemplate = tpl.Slug
	}
	title := RenderTemplate(titleTemplate, envelope.Variables)
	body := RenderTemplate(tpl.Body, envelope.Variables)

	payload := &PushPayload{
		Tokens:    activeTokens,
		Title:     title,
		Body:      body,
		Data:      toStringMap(envelope.Variables),
		Overrides: envelope.ProviderOverrides,
	}

	p.statusUpdater.MarkProcessing(ctx, envelope.RequestID)
	sendErr := retry.Do(ctx, p.retryCfg, func() error {
		results, err := p.fcm.Send(ctx, payload)
		if err != nil {
			p.logger.Warn("fcm send failed", slog.Any("error", err), slog.String("request_id", envelope.RequestID))
			return err
		}
		return p.handleResults(ctx, envelope, results)
	})

	if sendErr != nil {
		p.metrics.IncFailed()
		p.statusUpdater.MarkFailed(ctx, envelope.RequestID, p.fcm.Name(), sendErr.Error())
		return sendErr
	}

	p.metrics.IncDelivered()
	return nil
}

func (p *PushProcessor) filterTokens(ctx context.Context, tokens []models.PushToken) ([]models.PushToken, error) {
	if len(tokens) == 0 {
		return nil, nil
	}
	filtered := make([]models.PushToken, 0, len(tokens))
	for _, token := range tokens {
		if token.Token == "" {
			continue
		}
		if !supportsFCM(token.Platform) {
			continue
		}
		if p.cache != nil {
			suppressed, err := p.cache.IsTokenSuppressed(ctx, token.Token)
			if err != nil {
				return nil, err
			}
			if suppressed {
				continue
			}
		}
		filtered = append(filtered, token)
	}
	return filtered, nil
}

func (p *PushProcessor) handleResults(ctx context.Context, envelope *models.MessageEnvelope, results []models.PushResult) error {
	if len(results) == 0 {
		return fmt.Errorf("fcm returned no results")
	}

	var failures []string
	for _, res := range results {
		if res.Status == models.ResultDelivered {
			continue
		}
		failures = append(failures, fmt.Sprintf("%s:%s", res.Token, res.Error))
		if p.cache != nil && isTokenFatal(res.Error) {
			_ = p.cache.SuppressToken(ctx, res.Token, 0)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("failed tokens: %s", strings.Join(failures, ", "))
	}

	p.statusUpdater.MarkDelivered(ctx, envelope.RequestID, p.fcm.Name())
	return nil
}

func supportsFCM(platform string) bool {
	switch strings.ToLower(platform) {
	case "android", "ios":
		return true
	default:
		return false
	}
}

func isTokenFatal(err string) bool {
	switch err {
	case "NotRegistered", "InvalidRegistration", "MismatchSenderId", "MessageTooBig":
		return true
	default:
		return false
	}
}

func toStringMap(vars map[string]interface{}) map[string]string {
	result := make(map[string]string, len(vars))
	for k, v := range vars {
		result[k] = fmt.Sprint(v)
	}
	return result
}

func localeFromEnvelope(envelope *models.MessageEnvelope) string {
	if envelope.Template.Locale != "" {
		return envelope.Template.Locale
	}
	return envelope.User.Locale
}
