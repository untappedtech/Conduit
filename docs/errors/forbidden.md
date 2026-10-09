# 403 Forbidden

The authenticated client does not have sufficient role permissions to perform the requested action.

---

## What This Error Means

Authentication succeeded, but the caller's role does not allow this operation. For example, a client with a `readOnly` role attempting a `POST` or `DELETE` request, or a `readWrite` client attempting to mutate the schema (`POST /v1/schema/<table>`).

---

## Common Causes

- Using a `readOnly` token to perform data insert, update, or delete operations.
- Using a `readWrite` token to create or drop tables (`/v1/schema`).
- Policy restrictions disallowing schema mutation without the `admin` role.

---

## How to Fix It

1. **Use an Appropriate Token Role:** Request a token associated with the required role:
    - For table data mutations: Ensure the token has the `readWrite` or `admin` role.
    - For schema creation or deletion: Ensure the token has the `admin` role.
2. **Review Policy Settings:** Verify `publicMutation` or `publicWrites` settings in your Conduit configuration.

---

## Example Request & Response

### Request

```http
DELETE /v1/books/42
Authorization: Bearer ro-token-12345
```

### Response

```json
{
    "error": {
        "status": 403,
        "title": "Forbidden",
        "message": "You do not have permission to perform this action.",
        "details": ["role 'readOnly' does not have write access to table 'books'"],
        "documentation": "https://conduit.untapped.tech/docs/errors/forbidden"
    }
}
```

---

[← Back to Error Index](index.md)
