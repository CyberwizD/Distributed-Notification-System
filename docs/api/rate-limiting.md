# Rate Limiting

The API Gateway enforces a simple token bucket backed by Redis:

- **Scope:** Client IP address (`rate_limit:{ip}`)
- **Limit:** 100 requests per minute (configurable via middleware)
- **Storage:** Redis `INCR` + `EXPIRE` pipeline to avoid race conditions

When the counter exceeds the limit the gateway returns:

```json
{
  "success": false,
  "message": "rate limit exceeded",
  "error": "E_RATE_LIMIT"
}
```

### Customising
- Adjust the limit/window inside `internal/middleware/rate_limit.go` or expose env vars.
- For authenticated clients you can swap the key to use `user_id` instead of IP.
- Distributed deployments only require all API Gateway instances to share the same Redis cluster.

### Monitoring
- Track Redis key TTLs and consider exporting rejected-request counts via the `/metrics` endpoint.
