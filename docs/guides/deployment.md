# Deployment Guide

> The repository bundles five services (API Gateway, Push worker, Email worker, User Service, Template Service). This guide explains how to bring them up locally or on a single VM using Docker Compose. For managed environments (Kubernetes, ECS, etc.) replicate the same environment variables and health/liveness checks.

---

## 1. Prerequisites

- Docker Engine 24+ and Docker Compose v2
- Node.js 20.x (needed for NestJS builds/migrations)
- Go 1.22+ (for gateway/push local dev)
- Python 3.11+ (for email service local dev)
- A `.env` file in the repo root (see committed template)

---

## 2. Configure Environment

1. Copy `.env` to your deploy host and update secrets:
   - `JWT_SECRET`
   - `INTERNAL_API_KEY` / `USER_SERVICE_INTERNAL_API_KEY`
   - `FCM_SERVER_KEY`, SMTP credentials, etc.
2. Ensure database/cache/broker URLs point to reachable infrastructure (or rely on the compose defaults).

---

## 3. Run Database Migrations

```bash
cd services/user_service && npx prisma migrate deploy
cd services/template_service && npx prisma migrate deploy
```

These commands provision the Postgres tables used by the NestJS services. The Go services auto-migrate their small tables on startup.

---

## 4. Build & Start the Stack

```bash
docker compose -f deploy/docker-compose.yml build
docker compose -f deploy/docker-compose.yml up -d
```

Service matrix:

| Service            | Image Context             | Port  |
|--------------------|---------------------------|-------|
| API Gateway (Go)   | `services/api_gateway`    | 8080  |
| Push Service (Go)  | `services/push_service`   | 8081  |
| Email Service (Py) | `services/email_service`  | 2525  |
| User Service (Nest)| `services/user_service`   | 3001  |
| Template Service   | `services/template_service` | 3000 |
| Redis, Postgres, RabbitMQ | `deploy/docker-compose.yml` | 6379 / 5432 / 5672 |

The compose file also mounts `deploy/scripts/init-rabbitmq.sh`, which creates `notifications.direct` with `email.queue`, `push.queue`, and `failed.queue`.

---

## 5. Seed Required Data

1. **Create a user** (with preferences + device tokens) via `POST http://localhost:3001/api/v1/users`.
2. **Create a template** via `POST http://localhost:3000/api/v1/templates` and activate at least one version.
3. **Obtain a JWT**: `POST /api/v1/auth/login` on the User Service, then use that token for the gateway.

---

## 6. Sanity Checks

| URL                              | Expectation                 |
|---------------------------------|-----------------------------|
| `GET http://localhost:8080/health` | API gateway healthy        |
| `GET http://localhost:8081/health` | Push worker healthy       |
| `GET http://localhost:2525/`       | Email service running     |
| `http://localhost:15672`          | RabbitMQ management (admin/admin123) |

Send a sample notification:
```bash
curl -X POST http://localhost:8080/v1/notifications/send \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
        "request_id": "00000000-0000-0000-0000-000000000001",
        "user_id": "<user-id>",
        "channel": "email",
        "template_slug": "welcome_email",
        "variables": { "name": "Tester" }
      }'
```

Check status:
```
curl http://localhost:8080/v1/notifications/00000000-0000-0000-0000-000000000001/status
```

---

## 7. Production Tips

- Place the API Gateway behind HTTPS (Ingress/WAF) and configure autoscaling.
- Use managed Postgres/Redis/RabbitMQ; store secrets in Vault/SSM instead of `.env`.
- Run multiple instances of the push/email workers and configure RabbitMQ `prefetch_count`.
- Enable GitHub Actions workflows (already provided) to publish GHCR images for each service.

With the above steps the complete distributed notification system is operational end-to-end.
