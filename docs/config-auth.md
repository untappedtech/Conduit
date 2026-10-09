# Authorization Overview

Conduit features a pluggable authorization architecture to protect your data. When authentication is enabled, requests are evaluated against an extensible chain of auth providers.

Clients authenticate by passing an API key or bearer token via HTTP headers or query parameters.

---

## Token Extraction

Conduit extracts credentials from requests in the following order:

1. **`Authorization` Header (Bearer Token):**
   ```http
   Authorization: Bearer <token>
   ```
2. **`Authorization` Header (Direct Token):**
   ```http
   Authorization: <token>
   ```
3. **`X-API-Key` Header:**
   ```http
   X-API-Key: <token>
   ```
4. **`api_key` or `token` Query Parameter:**
   ```
   GET /v1/books?token=<token>
   ```

---

## Permission Levels & Roles

Conduit defines three standard access roles:

| Role | Permitted HTTP Methods | Description |
| ---- | ---------------------- | ----------- |
| **`readOnly`** (`ro`) | `GET`, `HEAD` | Read access to table data and schema introspection. |
| **`readWrite`** (`rw`) | `GET`, `POST`, `PUT`, `PATCH`, `DELETE` | Full data management capabilities on table rows. |
| **`admin`** | All methods + Schema Mutation | Full data CRUD plus table creation (`POST /v1/schema/<table>`) and table deletion (`DELETE /v1/schema/<table>`). |

---

## Authorization Modes

Conduit supports three primary authorization strategies, which can be configured independently:

```yaml
auth:
  environmentEnabled: false
  dbEnabled: false
```

- [Environment Auth](config-auth-env.md) — Fast, static API tokens loaded from environment variables (`AUTH_TOKEN_ADMIN`, `AUTH_TOKEN_READ_WRITE`, `AUTH_TOKEN_READ_ONLY`).
- [Database Auth](config-auth-db.md) — Dynamic token lookup against a dedicated users or API keys database table, featuring high-speed in-memory LRU caching.
- [No-Op Auth](config-auth-noop.md) — Fallback mode when no auth providers are enabled or configured. All requests are permitted without authentication.

---

## Fallback & Chain Behavior

When both `environmentEnabled` and `dbEnabled` are `false`, Conduit defaults to **No-Op Auth**. In this mode, permissions are governed strictly by the global flags in `policy` (such as `publicReads`, `publicWrites`, and `publicMutation`).
