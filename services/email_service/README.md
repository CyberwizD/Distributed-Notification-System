# Email Service

Lightweight FastAPI service that accepts email send requests and delivers messages via SMTP. Designed to run locally or inside Docker and consume messages from RabbitMQ.

## Features
- Single and batch send endpoints
- Jinja2 templates (built-in)
- SMTP send with SSL / STARTTLS fallback
- RabbitMQ consumer integration
- Health endpoint for readiness checks

## Prerequisites
- Python 3.11+
- Docker & Docker Compose (for containerized runs)
- RabbitMQ (or use docker-compose included)

## Configuration (.env)
Place a `.env` in the `services/email_service` folder (do not commit). Example variables:

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=you@example.com
SMTP_PASSWORD=your-app-password
RABBITMQ_URL=amqp://admin:admin123@rabbitmq:5672/