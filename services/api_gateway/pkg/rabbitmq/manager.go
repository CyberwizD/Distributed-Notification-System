package rabbitmq

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/streadway/amqp"
)

// Manager maintains a single AMQP connection and helps declare topology.
type Manager struct {
	url    string
	conn   *amqp.Connection
	logger *slog.Logger
	mu     sync.RWMutex
}

func NewManager(url string, logger *slog.Logger) (*Manager, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Manager{
		url:    url,
		conn:   conn,
		logger: logger,
	}, nil
}

func (m *Manager) Connection() *amqp.Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.conn == nil {
		return nil
	}
	err := m.conn.Close()
	m.conn = nil
	return err
}

// DeclareNotificationTopology ensures exchange/queues exist before publishing.
func (m *Manager) DeclareNotificationTopology(exchange string, routing map[string]string, dlq string) error {
	ch, err := m.Connection().Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	if dlq != "" {
		if _, err := ch.QueueDeclare(
			dlq,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("declare dlq: %w", err)
		}
	}

	for queue, key := range routing {
		args := amqp.Table{}
		if dlq != "" {
			args["x-dead-letter-exchange"] = ""
			args["x-dead-letter-routing-key"] = dlq
		}

		if _, err := ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			args,
		); err != nil {
			return fmt.Errorf("declare queue %s: %w", queue, err)
		}

		if err := ch.QueueBind(
			queue,
			key,
			exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("bind queue %s: %w", queue, err)
		}
	}

	return nil
}
