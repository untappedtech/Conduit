# 404 Not Found

The requested table or record does not exist in the database.

---

## What This Error Means

Conduit attempted to locate a database table or a specific primary key record, but no matching entity was found.

---

## Common Causes

- The specified table does not exist in the database.
- The record primary key supplied in the URL (`/v1/<table>/<id>`) does not match any existing row.
- Mistyped URL path or incorrect base path.

---

## How to Fix It

1. **Check Table Existence:** Query `GET /v1/schema` to list all tables recognized by Conduit.
2. **Verify Record ID:** Confirm the record ID exists by listing records with `GET /v1/<table>?where=id = <id>`.
3. **Verify URL Path:** Ensure your request adheres to the configured base path (e.g. `/v1/users/1` instead of `/users/1`).

---

## Example Request & Response

### Request

```http
GET /v1/books/9999
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
