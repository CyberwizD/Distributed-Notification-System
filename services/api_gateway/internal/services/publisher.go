package services

import (
	"encoding/json"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/models"
	"github.com/streadway/amqp"
)

// Publisher publishes messages to RabbitMQ.
type Publisher struct {
	conn *amqp.Connection
}

// NewPublisher creates a new Publisher.
func NewPublisher(conn *amqp.Connection) *Publisher {
	return &Publisher{conn: conn}
}

// Publish publishes a message to the notifications.direct exchange.
func (p *Publisher) Publish(envelope *models.MessageEnvelope) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"notifications.direct", // exchange
		envelope.Channel,       // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return err
	}

	return nil
}
