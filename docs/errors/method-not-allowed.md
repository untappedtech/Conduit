# 405 Method Not Allowed

The HTTP method used in the request is not supported on the specified endpoint.

---

## What This Error Means

Conduit recognized the target endpoint path, but the HTTP verb (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`) is not allowed for that resource.

---

## Common Causes

- Sending a `PUT` or `PATCH` request to a collection endpoint (`/v1/<table>`) instead of a single record endpoint (`/v1/<table>/<id>`).
- Sending `DELETE` or `PUT` to `/v1/schema` without specifying a table name.
- Using unsupported methods such as `TRACE` or `CONNECT`.

---

## How to Fix It

1. **Verify Allowed Methods:**
   - Collection endpoint `/v1/<table>`: Supports `GET` and `POST`.
   - Record endpoint `/v1/<table>/<id>`: Supports `GET`, `PUT`, `PATCH`, and `DELETE`.
   - Schema catalog `/v1/schema`: Supports `GET`.
   - Table schema `/v1/schema/<table>`: Supports `GET`, `POST`, and `DELETE`.
2. **Review API Documentation:** Check the [CRUD API](crud-api.md) and [Schema API](schema-api.md) references.

---

## Example Request & Response

### Request
```http
PUT /v1/books
Content-Type: application/json

{"title": "Updated Title"}
```

### Response
```json
{
  "error": {
    "status": 405,
    "title": "Method Not Allowed",
    "message": "The HTTP method 'PUT' is not allowed on collection endpoint '/v1/books'.",
    "details": [
      "Use PUT on /v1/books/<id> to replace an individual record"
    ],
    "documentation": "https://conduit.untapped.tech/docs/errors/method-not-allowed"
  }
}
```

---

[← Back to Error Index](index.md)
