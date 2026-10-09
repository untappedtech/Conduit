# 409 Conflict

The request could not be completed due to a conflict with the current state of the database.

---

## What This Error Means

The database driver rejected an insert or update query because it violates a relational constraint, such as a unique column constraint or duplicate primary key.

---

## Common Causes

- Inserting a record with a primary key that already exists in the table.
- Inserting or updating a record that violates a `UNIQUE` constraint (e.g. duplicate email address).
- Concurrency conflicts during simultaneous updates.

---

## How to Fix It

1. **Verify Uniqueness:** Ensure values for unique columns (such as usernames, email addresses, or external IDs) are not already present in the database.
2. **Use Updates Instead of Inserts:** If the record already exists, use `PUT /v1/<table>/<id>` or `PATCH /v1/<table>/<id>` rather than `POST /v1/<table>`.
3. **Inspect Auto-Increment Keys:** For auto-incrementing tables, avoid explicitly providing the primary key field in `POST` payloads.

---

## Example Request & Response

### Request

```http
POST /v1/users
Content-Type: application/json

{
  "username": "kyle",
  "email": "kyle@example.com"
}
```

### Response

```json
{
    "error": {
        "status": 409,
        "title": "Conflict",
        "message": "Unique constraint violation occurred.",
        "details": ["UNIQUE constraint failed: users.email"],
        "documentation": "https://conduit.untapped.tech/docs/errors/conflict"
    }
}
```

---

[← Back to Error Index](index.md)
