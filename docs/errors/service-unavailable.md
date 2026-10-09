# 503 Service Unavailable

The database or upstream service is temporarily unavailable.

---

## What This Error Means

Conduit is running, but it cannot connect to the configured database backend or an essential subsystem has timed out.

---

## Common Causes

- Database server is offline, restarting, or unreachable over the network.
- Exhausted database connection pool.
- Scheduled maintenance window in progress on cloud database instances (e.g. Turso / libSQL, Azure SQL, AWS RDS).
- Database timeout during high traffic or connection lock contention.

---

## How to Fix It

1. **Verify Database Connectivity:** Ensure the database server is running and reachable from the Conduit host using tools like `ping`, `nc`, or standard database CLI tools (`psql`, `mysql`, `sqlite3`).
2. **Review Connection String (DSN):** Check hostnames, IP addresses, ports, and credentials in `config.yaml`.
3. **Retry with Backoff:** If the database is undergoing a brief restart, retry the request after a short interval.

---

## Example Request & Response

### Response

```json
{
    "error": {
        "status": 503,
        "title": "Service Unavailable",
        "message": "The database service is temporarily unavailable.",
        "details": ["connection refused to postgres:5432"],
        "documentation": "https://conduit.untapped.tech/docs/errors/service-unavailable"
    }
}
```

---

[← Back to Error Index](index.md)
