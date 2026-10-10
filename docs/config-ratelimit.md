# Rate Limiting Configuration

Conduit features an embedded, zero-dependency Token Bucket rate limiter powered by Go's `golang.org/x/time/rate` package. It operates at the HTTP server's outermost perimeter before request processing, authentication, or database query execution, protecting your REST endpoints from volumetric denial-of-service and brute-force exhaustion.

---

## Configuration Settings

```yaml
rate_limit:
    enabled: true
    requests_per_second: 100
    burst: 200
```

| Field                 | Type      | Default | Description                                                                                                                                           |
| --------------------- | --------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enabled`             | `boolean` | `false` | Master toggle for the Token Bucket rate limiting middleware.                                                                                          |
| `requests_per_second` | `float`   | `100`   | Sustained refill rate. Number of token credits replenished each second. Fractional rates (e.g. `0.5` for 1 request per 2 seconds) are supported.      |
| `burst`               | `integer` | `200`   | Maximum bucket capacity. Dictates the allowable burst of concurrent requests. If `<= 0`, automatically defaults to `int(requests_per_second)` or `1`. |

---

## Client Identification & Keying

The rate limiter evaluates each incoming request to establish a unique limiter bucket using a **Token-First, IP-Fallback** strategy:

```
                          ┌───────────────────────────┐
                          │     Incoming Request      │
                          └─────────────┬─────────────┘
                                        │
                         Bearer Token / API Key present?
                                       / \
                                 YES  /   \  NO
                                     v     v
                         ┌──────────────┐   ┌──────────────────────────┐
                         │  Key:        │   │ Resolve Client IP via:   │
                         │  token:<key> │   │ 1. X-Forwarded-For       │
                         └──────────────┘   │ 2. X-Real-IP             │
                                            │ 3. RemoteAddr            │
                                            └────────────┬─────────────┘
                                                         │
                                                         v
                                            ┌──────────────────────────┐
                                            │  Key: ip:<resolved-ip>   │
                                            └──────────────────────────┘
```

1. **Authenticated Clients (`token:<token>`):** If an authorization credential is provided via `Authorization: Bearer <token>`, `X-API-Key: <token>`, or the `?api_key=` query parameter, Conduit assigns the bucket based on the token value. This ensures users retain their quota across roaming networks and NAT gateways.
2. **Anonymous Clients (`ip:<client-ip>`):** When unauthenticated, Conduit keys the bucket to the client IP address.

### Reverse Proxy & Gateway IP Extraction

When Conduit is deployed behind load balancers or reverse proxies (such as NGINX, Cloudflare, AWS ALB, or Envoy), client IP addresses are resolved in order of precedence:

1. **`X-Forwarded-For`:** Conduit inspects the header and extracts the leftmost (original client) IP address from comma-separated lists.
2. **`X-Real-IP`:** Used if `X-Forwarded-For` is absent.
3. **`RemoteAddr`:** Direct TCP socket peer address (with port stripped) as fallback.

---

## Memory Management & Janitor Sweeper

Client limiters are stored in an in-memory, thread-safe hash map (`sync.Map`) with atomic nanosecond timestamps recording each client's last activity:

- **Background Sweeper:** When `rate_limit.enabled: true`, a background cleanup goroutine wakes up every **60 seconds**.
- **Stale Eviction:** Clients that have been inactive for more than **3 minutes** are automatically evicted from memory.
- **Graceful Termination:** The background goroutine shuts down cleanly when the Conduit HTTP server terminates.

---

## Rate Limit Exceeded Response

When a client exhausts their token bucket, Conduit immediately halts request execution and responds with HTTP `429 Too Many Requests`:

### HTTP Response Headers & Status

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json; charset=utf-8
```

### Response Payload

```json
{
  "error": "Rate limit exceeded.",
  "code": 429
}
```

The error response is serialized using Conduit's multi-format response encoder, respecting the client's `Accept` header or `?format=` parameter (JSON, XML, YAML, TOML, CBOR, NDJSON, or CSV).

For more details on resolving client-side rate limit errors, see [429 Too Many Requests](errors/too-many-requests.md).

