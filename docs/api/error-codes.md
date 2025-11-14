# Error Codes

| Code | HTTP Status | Description                                   | Typical Cause                         |
|------|-------------|-----------------------------------------------|---------------------------------------|
| `E_VALIDATION` | 400 | Payload validation failed                    | Missing request fields, invalid UUIDs |
| `E_UNAUTHORIZED` | 401 | Authentication header missing/invalid      | JWT expired or absent                 |
| `E_RATE_LIMIT` | 429 | Client exceeded gateway rate limit           | Too many requests per minute per IP   |
| `E_SERVICE_DOWN` | 502 | Upstream dependency unavailable            | User/Template service outage          |
| `E_QUEUE_PUBLISH` | 500 | Failed to enqueue notification             | RabbitMQ unreachable                  |
| `E_STATUS_LOOKUP` | 404 | Status not found                           | Unknown `request_id`                  |

These codes are wrapped inside the standard response envelope (`success`, `message`, `error`). Use the `X-Correlation-ID` header to trace failures across services.
