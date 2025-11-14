# Monitoring & Observability

Keeping an eye on a distributed notification system requires visibility into HTTP traffic (API gateway), queue depth (RabbitMQ), worker throughput (email/push services), and storage health (Postgres/Redis). This guide outlines the built-in touch points and how to wire them into a standard monitoring stack.

---

## 1. Metrics Endpoints

| Service        | Endpoint                  | Notes                                             |
|----------------|---------------------------|---------------------------------------------------|
| API Gateway    | `GET /metrics`            | Requests total/failed + avg latency               |
| Push Service   | `GET /metrics`            | Consumed / delivered / failed / retried counters  |
| RabbitMQ       | `http://:15672`           | Queue depth, consumers, publish/ack rates         |
| Email Service  | (planned) use Docker logs | SMTP success/failure currently logged only        |

Scrape these with Prometheus or another collector. For example:
```yaml
- job_name: api-gateway
  static_configs:
    - targets: ['api_gateway:8080']
- job_name: push-service
  static_configs:
    - targets: ['push_service:8081']
```

---

## 2. Logging

### Correlation IDs
- API Gateway injects/propagates `X-Correlation-ID`.
- Include this ID in downstream logs (push/email workers already log `request_id` and `correlation_id`).

### Log Aggregation
- In Docker Compose use `docker compose logs -f <service>`.
- In production ship logs to ELK/Datadog/CloudWatch. Structure logs as JSON (Go code already uses `slog` which emits key/value pairs).

---

## 3. Health Checks

| Service          | Endpoint                     |
|------------------|------------------------------|
| API Gateway      | `GET /health`                |
| Push Service     | `GET /health`                |
| Email Service    | `GET /`                      |
| User Service     | `GET /api/v1/health`         |
| Template Service | `GET /api/v1/health`         |
| RabbitMQ         | `rabbitmqctl status` / `/health/checks/` |

Configure your orchestrator (Kubernetes, ECS, Nomad) to use these endpoints for liveness/readiness probes.

---

## 4. Alerting Suggestions

| Metric / Condition                 | Threshold                         | Action                                 |
|------------------------------------|-----------------------------------|----------------------------------------|
| RabbitMQ queue depth               | `email.queue` or `push.queue` > 1k for 5 min | Autoscale consumers / investigate provider outages |
| Notification failure rate          | `failed / delivered > 1%`         | Inspect provider overrides, down providers |
| API Gateway latency                | p95 > 500 ms                      | Inspect upstream dependencies (user/template services) |
| Redis / Postgres availability      | connection errors > 0             | Fallback or raise incident             |

Use Prometheus Alertmanager, Datadog monitors, or equivalent.

---

## 5. Tracing (Optional)

While not wired by default, the correlation ID pattern makes it easy to add tracing later:
1. Add OpenTelemetry instrumentation to the API Gateway (Gin middleware) and push/email services.
2. Forward traces to Jaeger/Tempo.
3. Pass `traceparent` headers when calling downstream HTTP services.

---

## 6. Runbook Snippets

**Drain stuck messages**
```bash
rabbitmqadmin get queue=email.queue requeue=false
```

**Inspect status store**
```sql
SELECT * FROM notification_statuses ORDER BY updated_at DESC LIMIT 20;
```

**Check push worker metrics**
```bash
curl http://localhost:8081/metrics
```

With these hooks you can observe throughput, latency, and failure scenarios across the entire distributed notification system. Augment them with provider-specific dashboards (FCM/SMTP) as you integrate real delivery channels.
