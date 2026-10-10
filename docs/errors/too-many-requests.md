# 429 Too Many Requests

The client has exceeded the permitted request rate limits.

---

## What This Error Means

Conduit's built-in Token Bucket rate limiter (configured under `rate_limit`) has determined that the client has depleted their token quota. Incoming requests are rejected at the server perimeter before database queries, authorization checks, or business logic execute.

---

## Common Causes

- Sustained request throughput exceeding the configured `requests_per_second`.
- Rapid bursts of concurrent requests exceeding the configured `burst` bucket capacity.
- Multiple corporate or office clients sharing a single egress IP address (NAT) without presenting an `Authorization` Bearer token or `X-API-Key`, causing them to share an `ip:<address>` rate limiter bucket.
- Automated scripts or crawlers running tight polling loops without backoff delays.

---

## How to Fix It

1. **Implement Exponential Backoff:** When receiving a `429` status, pause execution and retry with exponential backoff and randomized jitter.
2. **Authenticate Requests:** Supply a unique API token (`Authorization: Bearer <token>` or `X-API-Key: <token>`). Conduit keys rate limiters by token when available (`token:<token>`), preventing individual users from being throttled by shared IP quotas.
3. **Tune Rate Limits:** In high-throughput or trusted production environments, adjust `requests_per_second` and `burst` in your Conduit configuration file:
   ```yaml
   rate_limit:
       enabled: true
       requests_per_second: 500
       burst: 1000
   ```
4. **Disable Rate Limiting for Internal Services:** If Conduit is protected by an upstream API gateway or CDN that already enforces rate limits, set `rate_limit.enabled: false`.

---

## Example Request & Response

### Response Headers

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json; charset=utf-8
```

### Response Body

```json
{
  "error": "Rate limit exceeded.",
  "code": 429
}
```

---

[← Back to Error Index](index.md)
