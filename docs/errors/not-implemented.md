# 501 Not Implemented

The requested action or database feature is not implemented by the configured database driver.

---

## What This Error Means

The endpoint or functionality requested is part of Conduit's specification, but the active database engine driver does not support it (for example, dynamic DDL table creation on specialized columnar engines or unsupported constraints).

---

## Common Causes

- Attempting an operation not supported by the underlying driver (e.g. schema manipulation or specific constraint types not supported by ClickHouse or custom drivers).
- Future features defined in API specifications that are currently stubs.

---

## How to Fix It

1. **Check Driver Documentation:** Review the driver compatibility table in [Database Drivers](architecture-database-drivers.md).
2. **Use an Alternate Driver:** Switch to SQLite, PostgreSQL, MySQL, or SQL Server if dynamic DDL schema mutations are required.

---

## Example Request & Response

### Response
```json
{
  "error": {
    "status": 501,
    "title": "Not Implemented",
    "message": "This feature is not implemented for the configured database driver.",
    "details": [
      "driver 'clickhouse' does not support dynamic table schema drops"
    ],
    "documentation": "https://conduit.untapped.tech/docs/errors/not-implemented"
  }
}
```

---

[← Back to Error Index](index.md)
