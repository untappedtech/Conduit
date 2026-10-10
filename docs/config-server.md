# Server Configuration

The `server` block controls network binding, TCP listening ports, API route base paths, and query pagination limits. Top-level operational settings also manage interactive API documentation enablement (`enable_docs`), AST complexity limits (`max_query_depth`), and incoming throughput limits (`rate_limit`).

---

## Configuration Settings

```yaml
server:
    host: "0.0.0.0"
    port: 8080
    base_path: "/v1/"
    default_limit: 50

enable_docs: false
max_query_depth: 5

rate_limit:
    enabled: false
    requests_per_second: 100
    burst: 200
```

| Field                 | Type      | Default     | Description                                                                                                                                                                  |
| --------------------- | --------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `server.host`         | `string`  | `"0.0.0.0"` | Network interface IP address to bind to. Use `127.0.0.1` for local-only access or `0.0.0.0` for all available interfaces.                                                    |
| `server.port`         | `integer` | `8080`      | TCP port on which the HTTP server listens.                                                                                                                                   |
| `server.base_path`    | `string`  | `"/v1/"`    | Base URL prefix for all CRUD data routes, schema endpoints, and documentation routes. Automatically normalized with leading and trailing slashes.                          |
| `server.default_limit`| `integer` | `50`        | Default maximum rows returned by list endpoints (`GET /<base_path>/data/<table>`) when no `?limit=` query parameter is provided. Non-positive numbers fallback to `50`.    |
| `enable_docs`         | `boolean` | `false`     | When `true`, enables the embedded interactive Scalar documentation UI (`/docs`) and generated OpenAPI 3.0 specification (`/openapi.json`). When `false`, returns 404 Not Found. |
| `max_query_depth`     | `integer` | `5`         | Maximum allowed depth for nested parenthetical groups in `?where=` filter clauses to guard against AST complexity exhaustion. Values `<= 0` fallback to `5`.                 |
| `rate_limit`          | `object`  | *See table* | Token Bucket rate limiter configuration keyed by client IP or Bearer/API token. See [Rate Limiting](config-ratelimit.md) for full configuration details.                   |

---

## Base Path Configuration

By default, Conduit registers routes under `/v1/`:

- `GET|POST|PUT|PATCH|DELETE /v1/data/<table>[/:id]` (Dynamic CRUD endpoints)
- `GET|POST /v1/schema[/:table]` (Schema inspection and creation)
- `GET /v1/docs` (Interactive Scalar UI — requires `enable_docs: true`)
- `GET /v1/openapi.json` (OpenAPI Spec — requires `enable_docs: true`)

You can customize the base path to fit your reverse proxy or API gateway architecture:

```yaml
server:
    base_path: "/api/v2/"
```

Endpoints will then be served under:

- `http://localhost:8080/api/v2/data/<table>`
- `http://localhost:8080/api/v2/schema`
- `http://localhost:8080/api/v2/docs` (when `enable_docs: true`)
- `http://localhost:8080/api/v2/openapi.json` (when `enable_docs: true`)

> **Note:** When `enable_docs: true`, Conduit also provides root fallback aliases at `/docs` and `/openapi.json` regardless of `base_path` for convenience.

---

## Query Complexity Bounding (`max_query_depth`)

Conduit restricts nested parenthetical grouping inside `?where=` query parameters using `max_query_depth`. If a client sends a filter whose AST parenthetical nesting depth exceeds this value, Conduit immediately rejects the request with HTTP `400 Bad Request` before compiling SQL:

```json
{
  "error": "query too complex: nesting depth 6 exceeds maximum allowed depth of 5",
  "code": 400
}
```

---

## Pagination Limit Semantics

- **Default Limit (`default_limit`):** When a client requests `GET /v1/data/users` without specifying `limit`, Conduit applies `default_limit` (e.g. 50).
- **Unlimited Mode:** Clients can explicitly supply `?limit=0` to fetch all rows matching their query criteria (subject to database timeout and policy constraints).
- **Custom Limits:** Clients can specify any positive integer, e.g. `?limit=100`.
