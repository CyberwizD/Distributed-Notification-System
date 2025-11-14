package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/models"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/services"
	"github.com/streadway/amqp"
)

type PushConsumer struct {
	base          *BaseConsumer
	processor     *services.PushProcessor
	logger        *slog.Logger
	maxDeliveries int
}

func NewPushConsumer(base *BaseConsumer, processor *services.PushProcessor, logger *slog.Logger, maxDeliveries int) *PushConsumer {
	if maxDeliveries <= 0 {
		maxDeliveries = 5
	}
	return &PushConsumer{
		base:          base,
		processor:     processor,
		logger:        logger,
		maxDeliveries: maxDeliveries,
	}
}

func (p *PushConsumer) Start(ctx context.Context) error {
	return p.base.Start(ctx, p.handleDelivery)
}

func (p *PushConsumer) handleDelivery(ctx context.Context, msg amqp.Delivery) error {
	var envelope models.MessageEnvelope
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		p.logger.Error("failed to unmarshal envelope", slog.Any("error", err))
		_ = msg.Reject(false)
		return err
	}

	if err := p.processor.Process(ctx, &envelope); err != nil {
		requeue := p.shouldRetry(&msg)
		if requeue {
			p.logger.Warn("processing failed, message requeued", slog.String("request_id", envelope.RequestID), slog.Any("error", err))
		} else {
			p.logger.Error("processing failed, message dead-lettered", slog.String("request_id", envelope.RequestID), slog.Any("error", err))
		}
		_ = msg.Nack(false, requeue)
		return err
	}

	return msg.Ack(false)
}

func (p *PushConsumer) shouldRetry(msg *amqp.Delivery) bool {
	attempts := deliveryAttempts(msg)
	return attempts < p.maxDeliveries
}

func deliveryAttempts(msg *amqp.Delivery) int {
	if msg.Headers == nil {
		if msg.Redelivered {
			return 1
		}
		return 0
	}
	if raw, ok := msg.Headers["x-death"]; ok {
		if deaths, ok := raw.([]interface{}); ok && len(deaths) > 0 {
			if table, ok := deaths[0].(amqp.Table); ok {
				if count, ok := table["count"].(int64); ok {
					return int(count)
				}
			}
		}
	}
	if msg.Redelivered {
		return 1
	}
	return 0
}
