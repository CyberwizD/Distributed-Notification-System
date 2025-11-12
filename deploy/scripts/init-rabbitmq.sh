#!/bin/bash

# Wait for RabbitMQ to be ready
until rabbitmqctl status; do
  echo "Waiting for RabbitMQ to start..."
  sleep 5
done

# Declare the direct exchange
rabbitmqadmin declare exchange name=notifications.direct type=direct

# Declare the queues
rabbitmqadmin declare queue name=email.queue durable=true
rabbitmqadmin declare queue name=push.queue durable=true
rabbitmqadmin declare queue name=failed.queue durable=true

# Bind the queues to the exchange
rabbitmqadmin declare binding source="notifications.direct" destination_type="queue" destination="email.queue" routing_key="email"
rabbitmqadmin declare binding source="notifications.direct" destination_type="queue" destination="push.queue" routing_key="push"
rabbitmqadmin declare binding source="notifications.direct" destination_type="queue" destination="failed.queue" routing_key="failed"

echo "RabbitMQ setup complete."
