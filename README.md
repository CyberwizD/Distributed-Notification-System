# Distributed Notification System

> **Stage 4 Backend Task – HNG Internships**  
> Technology stack: Go, NestJS, Python, RabbitMQ, Redis, PostgreSQL, Docker

This repository contains a complete distributed notification system composed of five microservices plus supporting infrastructure. The goal is to accept notification requests via an API Gateway, enrich them with user and template data, enqueue them through RabbitMQ, process delivery via channel-specific workers (email/push), and track status end-to-end.

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Services](#services)
3. [Message Flow](#message-flow)
4. [Technology Stack](#technology-stack)
5. [Local Development](#local-development)
6. [Running with Docker Compose](#running-with-docker-compose)
7. [Environment Variables](#environment-variables)
8. [API Contracts & Documentation](#api-contracts--documentation)
9. [Monitoring & Observability](#monitoring--observability)
10. [CI/CD](#cicd)
11. [Testing](#testing)
12. [System Diagram](#system-diagram)

---

## Architecture Overview

```
Clients --> API Gateway (Go)
   |             |--> User Service (NestJS, Postgres, Redis)
   |             |--> Template Service (NestJS, Postgres)
   |             |--> Redis (rate limit / idempotency)
   |             '--> RabbitMQ notifications.direct exchange
   |
RabbitMQ queues: email.queue, push.queue, failed.queue (DLQ)
   |                              |
Email Worker (Python, SMTP)   Push Worker (Go, FCM)
   |                              |
Status Store (Postgres notification_statuses table)
```

Key characteristics:
- **Synchronous REST** for metadata enrichment (user preferences, template rendering).
- **Asynchronous messaging** for delivery (RabbitMQ direct exchange).
- **Idempotency & rate limiting** enforced via Redis.
- **Status tracking** persisted in Postgres and exposed via the API Gateway.

---

## Services

| Service            | Path                            | Language | Purpose                                                                  |
|--------------------|---------------------------------|----------|--------------------------------------------------------------------------|
| API Gateway        | `services/api_gateway`          | Go       | Entry point, validation, idempotency, Redis rate limiting, RabbitMQ pub. |
| Push Service       | `services/push_service`         | Go       | Consumes `push.queue`, fetches templates, sends via FCM, updates status. |
| Email Service      | `services/email_service`        | Python   | Consumes `email.queue`, renders templates, sends via SMTP, updates status|
| User Service       | `services/user_service`         | NestJS   | CRUD + authentication, stores preferences, exposes internal APIs.        |
| Template Service   | `services/template_service`     | NestJS   | Template CRUD, versioning, render endpoint.                              |

Supporting infrastructure lives under `deploy/` (Docker Compose, RabbitMQ init, health scripts) and `docs/` (OpenAPI specs, guides, diagrams).

---

## Message Flow

1. **Client → API Gateway**
   - `POST /v1/notifications/send`
   - Gateway validates payload, enforces rate limit, checks idempotency.
2. **Gateway → User Service**
   - Fetches preferences + device tokens via internal API (`X-Internal-API-Key`).
3. **Gateway → Template Service**
   - Retrieves the active template (subject/body) for the requested slug/locale.
4. **Gateway → RabbitMQ**
   - Publishes `MessageEnvelope` to `notifications.direct` with routing key `email` or `push`.
5. **Workers consume queues**
   - Email/Pull services render content, send via provider (SMTP/FCM), retry failures, dead-letter irrecoverable messages.
6. **Status updates**
   - Workers upsert into `notification_statuses` (Postgres). API Gateway exposes `GET /v1/notifications/:id/status`.

---

## Technology Stack

- **Go 1.22+**: API Gateway, Push worker
- **NestJS (Node.js 20)**: User & Template services
- **Python 3.11 (FastAPI)**: Email service
- **RabbitMQ**: Message broker (direct exchange)
- **Redis**: Rate limiting, idempotency cache
- **PostgreSQL**: Primary database for user/template data + shared status table
- **Docker / Docker Compose**: Local orchestration
- **GitHub Actions**: CI/CD pipelines per service

---

## Local Development

### Prerequisites
- Go 1.22+
- Node.js 20.x + npm
- Python 3.11 + pip
- Docker Desktop (for running dependencies)

### Setup
1. Clone repo and copy `.env` template (already provided) to your environment.
2. Start infrastructure (Redis/Postgres/RabbitMQ) via Docker:
   ```bash
   docker compose -f deploy/docker-compose.yml up redis postgres rabbitmq -d
   ```
3. Run each service in dev mode:
   - API Gateway: `cd services/api_gateway && go run ./cmd/server`
   - Push Service: `cd services/push_service && go run ./cmd/consumer`
   - Email Service: `cd services/email_service && uvicorn app.main:app --reload --port 2525`
   - User Service: `cd services/user_service && npm run start:dev`
   - Template Service: `cd services/template_service && npm run start:dev`

4. Run migrations:
   ```bash
   cd services/user_service && npx prisma migrate deploy
   cd services/template_service && npx prisma migrate deploy
   ```

5. Seed data (User Service to create users/preferences, Template Service to create templates).

---

## Running with Docker Compose

```bash
docker compose -f deploy/docker-compose.yml build
docker compose -f deploy/docker-compose.yml up
```

Once running:
- API Gateway: `http://localhost:8080`
- User Service Swagger: `http://localhost:3001/api/docs`
- Template Service Swagger: `http://localhost:3000/api/docs`
- RabbitMQ UI: `http://localhost:15672` (admin/admin123)
- Email Service test endpoint: `http://localhost:2525/test-email`
- Push Service: `http://localhost:8081/health`

---

## Environment Variables

See `.env` (root) for the full matrix. Key entries:

| Variable | Service | Description |
|----------|---------|-------------|
| `DATABASE_URL` | API Gateway / Push / Email | Postgres connection string |
| `USER_SERVICE_URL` + `USER_SERVICE_INTERNAL_API_KEY` | API Gateway | Internal user profile API |
| `TEMPLATE_SERVICE_URL` | Gateway / Workers | Template fetch URL |
| `RABBITMQ_URL` | Gateway / Workers | RabbitMQ connection |
| `REDIS_URL` | Gateway / Push | Rate limit + cache |
| `JWT_SECRET`, `JWT_EXPIRES_IN` | User Service | Auth token settings |
| `FCM_SERVER_KEY` | Push Service | Firebase Cloud Messaging |
| `SMTP_*` | Email Service | SMTP credentials |

All defaults are set for the Docker Compose network; override for production.

---

## API Contracts & Documentation

- **OpenAPI Specs** (YAML):
  - `docs/openapi/api-gateway.yml`
  - `docs/openapi/user-service.yml`
  - `docs/openapi/template-service.yml`
  - `docs/openapi/email-service.yml`
  - `docs/openapi/push-service.yml`
- **Response/Envelope Schemas**: `shared/schemas/`
- **Message Protobuf**: `shared/proto/notification.proto`
- **Guides**: `docs/api/` (authentication, contracts, error codes, rate limiting)

Use these files to generate Postman collections, swagger UIs, or SDKs.

---

## Monitoring & Observability

- **Metrics**: API Gateway & Push Service expose `/metrics` (Prometheus text format).
- **Health Checks**:
  - API Gateway: `GET /health`
  - Push Service: `GET /health`
  - Email Service: `GET /`
  - User/Template: `GET /api/v1/health`
- **Logs**: All services include correlation IDs; tail with `docker compose logs -f`.
- **RabbitMQ**: Monitor queue depth and consumer counts via the management UI.
- Detailed instructions: `docs/guides/monitoring.md`

---

## CI/CD

GitHub Actions workflows live under `.github/workflows/`:

| Workflow                | Description                                         |
|-------------------------|-----------------------------------------------------|
| `api-gateway.yml`       | Go build/test, Docker build/push                    |
| `push-service.yml`      | Go build/test, Docker build/push                    |
| `email-service.yml`     | Python checks, Docker build/push                    |
| `user-service.yml`      | npm install/test/build, Docker build/push           |
| `template-service.yml`  | npm install/test/build, Docker build/push           |

On `main`, images are pushed to GHCR (`ghcr.io/<owner>/<service>`).

---

## Testing

- **Go services**: `go test ./...`
- **NestJS services**: `npm run test -- --passWithNoTests`
- **Email service**: Add tests under `app/tests` and run `pytest`.
- Integration testing: use the provided `scripts/load-test.js` (k6) to simulate 1k req/min.

---

## System Diagram

Refer to `docs/diagrams/notification-system.drawio` for the visual representation covering:
- Service interactions, queue bindings, DLQ routing
- Retry/failure flows
- Database relationships
- Scaling considerations (stateless gateway/workers, shared infra)

Open the file using [draw.io](https://app.diagrams.net/) or import into Lucidchart/Miro.

---

## Contributing / Extending

1. Fork & clone the repo.
2. Create feature branches (`feature/<name>`).
3. Ensure `go test`, `npm run test`, or `pytest` pass where relevant.
4. Update OpenAPI docs & shared schemas when contract changes occur.
5. Open a PR with a detailed summary referencing any relevant docs or swagger updates.

---

## License

This project is provided for the HNG Internship Stage 4 task. Adapt/extend as needed for educational purposes. For commercial use, please contact the authors.
