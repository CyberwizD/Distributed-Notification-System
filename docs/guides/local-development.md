# Local Development

1. **Install dependencies**
   - Go 1.22+, Node.js 20.x, Python 3.11+
   - `npm ci` inside `services/user_service` and `services/template_service`
   - `pip install -r requirements.txt` in `services/email_service`

2. **Environment**
   - Copy `.env` and fill the secrets (`JWT_SECRET`, `INTERNAL_API_KEY`, etc.)
   - Start Redis/Postgres/RabbitMQ via Docker: `docker compose -f deploy/docker-compose.yml up redis postgres rabbitmq`

3. **Service-specific commands**
   - API Gateway: `cd services/api_gateway && go run ./cmd/server`
   - Push Service: `cd services/push_service && go run ./cmd/consumer`
   - User Service: `cd services/user_service && npm run start:dev`
   - Template Service: `cd services/template_service && npm run start:dev`
   - Email Service: `cd services/email_service && uvicorn app.main:app --reload --port 2525`

4. **Hot Reload**
   - Go services can use `air` or `reflex` if you prefer live reload.
   - NestJS services already ship with `--watch`.

5. **Testing**
   - Go: `go test ./...`
   - NestJS: `npm run test -- --watch`
   - Python: `pytest` (add tests under `services/email_service/tests`)

6. **Sample request**
```bash
curl -X POST http://localhost:8080/v1/notifications/send \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{ "request_id":"<uuid>", "user_id":"<uuid>", "channel":"push", "template_slug":"welcome_push", "variables":{"name":"Dev"} }'
```
