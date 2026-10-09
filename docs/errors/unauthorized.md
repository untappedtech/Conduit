# 401 Unauthorized

Authentication credentials are missing, malformed, or invalid for the requested operation.

---

## What This Error Means

The endpoint or operation requires authentication because public access has been disabled in the `policy` configuration (`publicReads: false`, `publicWrites: false`, or `publicMutation: false`), and no valid authorization token was provided.

---

## Common Causes

- Missing `Authorization` or `X-API-Key` header.
- Expired or revoked API key or bearer token.
- Typo in the authorization header or token value.
- Token does not exist in the environment or database auth store.

---

## How to Fix It

1. **Provide a Valid Token:** Include your API token using the `Authorization: Bearer <token>` or `X-API-Key: <token>` header.
2. **Verify Configuration:** Check whether environment tokens (`AUTH_TOKEN_*`) are correctly set in the server environment.
3. **Verify Database Auth Table:** For database-backed authentication, ensure the token exists and is active in the configured table.

---

## Example Request & Response

### Request

```http
POST /v1/books
Content-Type: application/json

{"title": "Dune"}
```

### Response

```json
{
    "error": {
        "status": 401,
        "title": "Unauthorized",
        "message": "Authentication is required to access this resource.",
        "details": ["missing or invalid authorization credentials"],
        "documentation": "https://conduit.untapped.tech/docs/errors/unauthorized"
    }
}
```

---

[← Back to Error Index](index.md)
