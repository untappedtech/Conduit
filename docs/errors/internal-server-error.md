# 500 Internal Server Error

An unexpected error occurred within the Conduit server or underlying database driver.

---

## What This Error Means

Conduit encountered an unhandled condition or lower-level system fault while processing the request.

---

## Common Causes

- Underlying database connection dropped or database crashed.
- Disk full, read-only filesystem, or disk I/O errors on SQLite or temporary stores.
- Database driver panic or unexpected nil reference.
- Low-level serialization or query execution failure.

---

## How to Fix It

1. **Inspect Server Logs:** Check the Conduit standard output logs (`log.Printf`) for the exact internal error message and stack trace.
2. **Check Database Health:** Ensure PostgreSQL, MySQL, SQL Server, ClickHouse, or Oracle server processes are running and accepting TCP connections.
3. **Verify Disk Space & Permissions:** For file-based SQLite databases, ensure the server process has read and write permissions to the database file and directory.
4. **Report Bugs:** If the issue is reproducible with a specific dataset or query, submit an issue on GitHub.

---

## Example Request & Response

### Response
```json
{
  "error": {
    "status": 500,
    "title": "Internal Server Error",
    "message": "An internal error occurred while processing the request.",
    "details": [
      "failed to execute statement: database disk image is malformed"
    ],
    "documentation": "https://conduit.untapped.tech/docs/errors/internal-server-error"
  }
}
```

---

[← Back to Error Index](index.md)
