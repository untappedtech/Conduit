# Server Configuration

The `server` section controls network binding, TCP listening ports, API route base paths, and query pagination limits.

---

## Configuration Settings

```yaml
server:
    host: "0.0.0.0"
    port: 8080
    base_path: "/v1/"
    default_limit: 50
```

| Field           | Type      | Default     | Description                                                                                                                                                        |
| --------------- | --------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `host`          | `string`  | `"0.0.0.0"` | Network interface IP address to bind to. Use `127.0.0.1` for local-only access or `0.0.0.0` for all available interfaces.                                          |
| `port`          | `integer` | `8080`      | TCP port on which the HTTP server listens.                                                                                                                         |
| `base_path`     | `string`  | `"/v1/"`    | Base URL prefix for all CRUD, schema, and documentation routes. Automatically normalized with leading and trailing slashes.                                        |
| `default_limit` | `integer` | `50`        | Default maximum rows returned by list endpoints (`GET /<base_path>/<table>`) when no `?limit=` query parameter is provided. Non-positive numbers fallback to `50`. |

---

## Base Path Configuration

By default, Conduit registers routes under `/v1/`:

- `GET /v1/<table>`
- `GET /v1/schema`
- `GET /v1/docs` (Interactive Scalar UI)
- `GET /v1/openapi.json` (OpenAPI Spec)

You can customize the base path to fit your reverse proxy or API gateway architecture:

```yaml
server:
    base_path: "/api/v2/"
```

Endpoints will then be served under:

- `http://localhost:8080/api/v2/<table>`
- `http://localhost:8080/api/v2/schema`
- `http://localhost:8080/api/v2/docs`

> **Note:** Conduit also provides root fallback aliases at `/docs` and `/openapi.json` regardless of `base_path` for convenience.

---

## Pagination Limit Semantics

- **Default Limit (`default_limit`):** When a client requests `GET /v1/users` without specifying `limit`, Conduit applies `default_limit` (e.g. 50).
- **Unlimited Mode:** Clients can explicitly supply `?limit=0` to fetch all rows matching their query criteria (subject to database timeout and policy constraints).
- **Custom Limits:** Clients can specify any positive integer, e.g. `?limit=100`.
