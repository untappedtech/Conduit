# 404 Not Found

The requested table or record does not exist in the database.

---

## What This Error Means

Conduit attempted to locate a database table or a specific primary key record, but no matching entity was found.

---

## Common Causes

- The specified table does not exist in the database.
- The record primary key supplied in the URL (`/v1/data/<table>/<id>`) does not match any existing row.
- Requesting `/docs` or `/openapi.json` when `enable_docs` is `false` in the configuration.
- Mistyped URL path or missing the `/data/` prefix on CRUD requests.

---

## How to Fix It

1. **Check Table Existence:** Query `GET /v1/schema` to list all tables recognized by Conduit.
2. **Verify Record ID:** Confirm the record ID exists by listing records with `GET /v1/data/<table>?where=id = <id>`.
3. **Verify URL Path:** Ensure your request adheres to the configured base path and `/data/` prefix (e.g. `/v1/data/users/1` instead of `/v1/users/1`).
4. **Enable Documentation:** If `/docs` or `/openapi.json` returns 404, set `enable_docs: true` in your Conduit configuration file.

---

## Example Request & Response

### Request

```http
GET /v1/data/books/9999
```

### Response

```json
{
    "error": {
        "status": 404,
        "title": "Not Found",
        "message": "The requested resource does not exist.",
        "details": ["Table 'books' record with id '9999' not found"],
        "documentation": "https://conduit.untapped.tech/docs/errors/not-found"
    }
}
```

---

[← Back to Error Index](index.md)
