# Distributed Notification System – Service Contracts

## Message Envelope
All asynchronous notifications share the same envelope. The authoritative schema lives in `shared/proto/notification.proto` and `shared/schemas/message-envelope.json`.

```json
{
  "request_id": "uuid",
  "correlation_id": "uuid-or-trace",
  "created_at": "2025-11-10T12:00:00Z",
  "channel": "email | push",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "locale": "en",
    "push_tokens": [
      { "token": "abc", "platform": "ios", "provider": "fcm" }
    ]
  },
  "template": {
    "slug": "welcome_email",
    "locale": "en",
    "version": 0,
    "subject": "Welcome {{name}}",
    "body": "<p>Hello {{name}}</p>"
  },
  "variables": { "name": "Wisdom" },
  "provider_overrides": {},
  "retry_count": 0
}
```

## HTTP Response Envelope
Every public API responds with the wrapper defined in `shared/schemas/response-wrapper.json`.

```json
{
  "success": true,
  "message": "notification queued",
  "data": {
    "request_id": "uuid",
    "status": "queued"
  },
  "meta": null
}
```

## Service Endpoints

| Service          | Base URL (in Docker)              | Key Endpoints                                                                                 |
|------------------|-----------------------------------|-----------------------------------------------------------------------------------------------|
| API Gateway      | `http://api_gateway:8080/v1`      | `POST /notifications/send`, `GET /notifications/:id/status`, `GET /health`, `GET /metrics`    |
| User Service     | `http://user_service:3001/api/v1` | `GET /users/:id/preferences`, `GET /internal/users/:id/notification-profile` (API-key only)   |
| Template Service | `http://template_service:3000/api/v1` | `GET /templates/:code/active?locale=en`, `POST /templates/render`                          |
| Push Service     | `http://push_service:8081/health` | Consumes `push.queue`, publishes status changes                                               |
| Email Service    | `http://email_service:2525/`      | `/send-email` for direct tests, consumes `email.queue` in production                          |

## Status Storage
All workers upsert into the shared `notification_statuses` table:

```
request_id TEXT PRIMARY KEY
status     TEXT
provider   TEXT
detail     TEXT
updated_at TIMESTAMPTZ
```

API Gateway writes `queued`. Consumers update to `processing`, `delivered`, or `failed`.

## Security
- Service-to-service calls into the User Service must include `X-Internal-API-Key` (see `INTERNAL_API_KEY` in docker-compose).
- Client requests to the API Gateway must include `Authorization: Bearer <token>` (middleware validates presence) and `X-Correlation-ID` (auto-generated if absent).

## Run Book
1. `docker compose -f deploy/docker-compose.yml up --build`
2. Seed templates/users (see respective README files).
3. Hit `POST http://localhost:8080/v1/notifications/send` with the sample payload from the README.
4. Monitor RabbitMQ at `http://localhost:15672`, API metrics at `/metrics`, and statuses via `GET /v1/notifications/:id/status`.
