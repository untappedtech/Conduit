# Environment Authorization

Environment-based authorization validates incoming API requests using static security tokens injected via environment variables.

This mode is ideal for containerized deployments (Docker, Kubernetes), staging environments, microservices, and automated CI/CD runners where managing a database table for API tokens is unnecessary.

---

## Configuration Settings

To enable environment authorization, set `environmentEnabled: true` under `auth`:

```yaml
auth:
    environmentEnabled: true
    dbEnabled: false
```

---

## Environment Variables

When `environmentEnabled` is `true`, Conduit reads authorization tokens from the following standard environment variables upon startup:

| Environment Variable    | Role Granted | Access Capabilities                                       |
| ----------------------- | ------------ | --------------------------------------------------------- |
| `AUTH_TOKEN_ADMIN`      | `admin`      | Full read/write access plus schema creation and deletion. |
| `AUTH_TOKEN_READ_WRITE` | `readWrite`  | Read, insert, update, patch, and delete data records.     |
| `AUTH_TOKEN_READ_ONLY`  | `readOnly`   | Read-only access to records and schema introspection.     |

---

## Example Usage

### 1. Set Environment Variables

```bash
# Linux / macOS
export AUTH_TOKEN_ADMIN="super-secret-admin-token"
export AUTH_TOKEN_READ_WRITE="rw-service-token-1234"
export AUTH_TOKEN_READ_ONLY="ro-dashboard-token-5678"

# Windows PowerShell
$env:AUTH_TOKEN_ADMIN="super-secret-admin-token"
$env:AUTH_TOKEN_READ_WRITE="rw-service-token-1234"
$env:AUTH_TOKEN_READ_ONLY="ro-dashboard-token-5678"
```

### 2. Start Conduit

```bash
conduit --config config.yaml
```

### 3. Make Authorized Requests

```bash
# Read-Only Query using Bearer Token
curl -H "Authorization: Bearer ro-dashboard-token-5678" http://localhost:8080/v1/players

# Data Write using Header
curl -X POST http://localhost:8080/v1/players \
  -H "Authorization: Bearer rw-service-token-1234" \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "score": 90}'

# Schema Mutation using Admin Token
curl -X POST http://localhost:8080/v1/schema/teams \
  -H "Authorization: Bearer super-secret-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"columns": [{"name": "id", "type": "integer", "pk": true}]}'
```
