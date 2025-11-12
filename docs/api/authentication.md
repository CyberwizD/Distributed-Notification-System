# Authentication

Every public entry point (API Gateway, User Service, Template Service) requires JWT bearer tokens issued by the User Service. Internal service-to-service hops rely on an API key (`X-Internal-API-Key`) so downstream services can reject untrusted callers.

---

## 1. Client → API Gateway

| Header             | Required | Description                                                                           |
|--------------------|----------|---------------------------------------------------------------------------------------|
| `Authorization`    | Yes      | `Bearer <JWT>` issued by the User Service `/api/v1/auth/login` endpoint               |
| `X-Correlation-ID` | Optional | Trace identifier. If omitted, the gateway generates a UUID and echoes it back         |

* Tokens expire per the `JWT_EXPIRES_IN` environment variable (default `1d`).
* Signing secret rotates via `JWT_SECRET`; update both the User Service and any token verifiers.

### Obtaining a Token
```bash
curl -X POST http://localhost:3001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@notify.io","password":"changeme"}'
```
Use the returned `access_token` when calling the API Gateway.

---

## 2. Service → Service (Internal)

Downstream Nest services expose internal-only routes guarded by `InternalApiGuard`. Gateway requests must include:

```
X-Internal-API-Key: <shared-secret>
```

Configure the key via `INTERNAL_API_KEY` (User Service) and `USER_SERVICE_INTERNAL_API_KEY` (API Gateway). Rotate it periodically and never expose it to clients.

---

## 3. RabbitMQ + Workers

Workers authenticate with RabbitMQ via the URL in `RABBITMQ_URL` (default `admin/admin123`). Production deployments should create per-service users with least-privilege permissions.

---

## 4. Operational Tips

1. Terminate TLS at the edge and forward only HTTPS traffic to the API Gateway.
2. Always propagate/ log the `X-Correlation-ID` so traces across microservices can be correlated.
3. Monitor token issuance & rejection metrics to detect abuse.
4. Store secrets (JWT key, internal API key, RabbitMQ creds, SMTP/FCM credentials) in a vault or secret manager rather than plain `.env` files in production.
