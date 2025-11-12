package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/streadway/amqp"
)

// BaseConsumer wires RabbitMQ connectivity, queue declaration and worker handling.
type BaseConsumer struct {
	conn         *amqp.Connection
	queue        string
	dlq          string
	prefetch     int
	workerCount  int
	logger       *slog.Logger
	exchangeName string
}

func NewBaseConsumer(conn *amqp.Connection, queue, dlq string, prefetch, workerCount int, logger *slog.Logger) *BaseConsumer {
	if prefetch <= 0 {
		prefetch = 50
	}
	if workerCount <= 0 {
		workerCount = 5
	}
	return &BaseConsumer{
		conn:         conn,
		queue:        queue,
		dlq:          dlq,
		prefetch:     prefetch,
		workerCount:  workerCount,
		logger:       logger,
		exchangeName: "notifications.direct",
	}
}

func (c *BaseConsumer) Start(ctx context.Context, handler func(context.Context, amqp.Delivery) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := c.setupQueue(ch); err != nil {
		return fmt.Errorf("queue setup failed: %w", err)
	}

	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return fmt.Errorf("qos configuration failed: %w", err)
	}

	deliveries, err := ch.Consume(
		c.queue,
		"",
		false, // autoAck
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < c.workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-deliveries:
					if !ok {
						return
					}
					if err := handler(ctx, msg); err != nil {
						c.logger.Error("handler returned error", slog.Any("error", err))
					}
				}
			}
		}(i)
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (c *BaseConsumer) setupQueue(ch *amqp.Channel) error {
	args := amqp.Table{}
	if c.dlq != "" {
		args["x-dead-letter-exchange"] = ""
		args["x-dead-letter-routing-key"] = c.dlq
	}

	if err := ch.ExchangeDeclare(
		c.exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := ch.QueueDeclare(
		c.queue,
		true,
		false,
		false,
		false,
		args,
	); err != nil {
		return err
	}

	if err := ch.QueueBind(
		c.queue,
		"push",
		c.exchangeName,
		false,
		nil,
	); err != nil {
		return err
	}

	if c.dlq != "" {
		if _, err := ch.QueueDeclare(
			c.dlq,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return err
		}
	}
	return nil
}
