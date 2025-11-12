package services

import (
	"context"

	"log/slog"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/repository"
)

const (
	StatusProcessing = "processing"
	StatusDelivered  = "delivered"
	StatusFailed     = "failed"
)

type StatusUpdater struct {
	store  *repository.StatusStore
	logger *slog.Logger
}

func NewStatusUpdater(store *repository.StatusStore, logger *slog.Logger) *StatusUpdater {
	return &StatusUpdater{
		store:  store,
		logger: logger,
	}
}

func (s *StatusUpdater) MarkProcessing(ctx context.Context, requestID string) {
	if err := s.store.UpdateStatus(ctx, requestID, StatusProcessing, "", ""); err != nil {
		s.logger.Error("failed to update processing status", slog.String("request_id", requestID), slog.Any("error", err))
	}
}

func (s *StatusUpdater) MarkDelivered(ctx context.Context, requestID, provider string) {
	if err := s.store.UpdateStatus(ctx, requestID, StatusDelivered, provider, ""); err != nil {
		s.logger.Error("failed to update delivered status", slog.String("request_id", requestID), slog.Any("error", err))
	}
}

func (s *StatusUpdater) MarkFailed(ctx context.Context, requestID, provider, detail string) {
	if err := s.store.UpdateStatus(ctx, requestID, StatusFailed, provider, detail); err != nil {
		s.logger.Error("failed to update failed status", slog.String("request_id", requestID), slog.Any("error", err))
	}
}
