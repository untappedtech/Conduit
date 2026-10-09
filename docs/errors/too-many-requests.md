# 429 Too Many Requests

The client has exceeded the permitted rate limits.

---

## What This Error Means

Too many requests have been received from the client within a configured time window.

---

## Common Causes

- Aggressive automated scraping or rapid sequential requests without rate limiting.
- Infinite client retry loops.
- Concurrent requests exhausting connection worker pools.

---

## How to Fix It

1. **Implement Exponential Backoff:** Delay subsequent requests and retry with exponential backoff and jitter.
2. **Batch Operations:** Where possible, combine multiple record inserts into batch requests rather than issuing one HTTP call per row.
3. **Cache Responses:** Use client-side caching for frequently read catalog data.

---

## Example Request & Response

### Response

```json
{
    "error": {
        "status": 429,
        "title": "Too Many Requests",
        "message": "Rate limit exceeded. Please slow down your requests.",
        "details": ["Retry after 30 seconds"],
        "documentation": "https://conduit.untapped.tech/docs/errors/too-many-requests"
    }
}
```

---

[← Back to Error Index](index.md)
