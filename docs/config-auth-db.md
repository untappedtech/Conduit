# Database Authorization

Database-backed authorization validates incoming API tokens against a dedicated table stored in a SQL database.

This mode is suited for multi-user applications, SaaS platforms, and enterprise setups where API keys and user tokens must be provisioned, rotated, and revoked dynamically at runtime.

---

## Configuration Settings

To enable database auth, set `dbEnabled: true` and configure the `dbAuth` parameters:

```yaml
auth:
  environmentEnabled: false
  dbEnabled: true
  dbAuth:
    driver: "sqlite"
    dsn: "./auth.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
    table: "api_keys"
    tokenColumn: "token"
    roleColumn: "role"
    roles:
      admin: "admin"
      readWrite: "rw"
      readOnly: "ro"
    cache:
      capacity: 10000
      ttlSeconds: 300
```

| Field | Type | Description |
| ----- | ---- | ----------- |
| `driver` | `string` | Database driver for the auth database (e.g. `sqlite`, `postgres`, `mysql`). Can point to the same DB or an isolated security database. |
| `dsn` | `string` | Connection DSN string for the authentication database. |
| `table` | `string` | Name of the table storing credentials. |
| `tokenColumn` | `string` | Column containing the secret API key or token string. |
| `roleColumn` | `string` | Column containing the user or key's role identifier. |
| `roles.admin` | `string` | Value in `roleColumn` that maps to Conduit's `admin` role. |
| `roles.readWrite` | `string` | Value in `roleColumn` that maps to Conduit's `readWrite` role. |
| `roles.readOnly` | `string` | Value in `roleColumn` that maps to Conduit's `readOnly` role. |
| `cache.capacity` | `integer` | Maximum number of authenticated tokens cached in memory (LRU). Default: `10000`. |
| `cache.ttlSeconds` | `integer` | Time-to-live in seconds for cached tokens before re-querying the database. Default: `300` (5 minutes). |

---

## Example Database Schema

You can store tokens in an SQLite, PostgreSQL, or MySQL table like this:

```sql
CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    role TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO api_keys (token, role, description) VALUES
    ('adm_sec_999888', 'admin', 'Primary admin token'),
    ('rw_sec_111222', 'rw', 'ETL sync worker'),
    ('ro_sec_333444', 'ro', 'Public dashboard');
```

---

## High-Performance Caching

To prevent querying the authentication database on every single incoming HTTP request, Conduit includes an internal, thread-safe LRU cache:
- Token lookups are cached up to `capacity` entries.
- Cached tokens automatically expire after `ttlSeconds`.
- Cache invalidation occurs automatically when expired, ensuring revoked tokens take effect cleanly.
