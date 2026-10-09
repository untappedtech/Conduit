# 400 Bad Request

The request could not be processed due to invalid parameters or malformed request syntax.

---

## What This Error Means

Conduit encountered a condition in the request preventing execution before touching the database. This frequently occurs when query parameters, column names, or payload structures fail syntactic validation.

---

## Common Causes

- Unknown column specified in `?order=` (e.g. `?order=non_existent_column:asc`).
- Unknown column referenced in `?where=` filter expressions.
- Missing required query parameters or malformed parameter values.
- Supplying invalid data types that cannot be parsed.

---

## How to Fix It

1. **Verify Column Names:** Ensure that columns referenced in `where` and `order` parameters match the schema of the target table. Check `GET /v1/schema/<table>` for the exact column names.
2. **Review Query Syntax:** Verify sorting direction syntax (`column` or `column:asc` or `column:desc`).
3. **Inspect Server Response Details:** Check the `details` field in the response envelope for the specific offending parameter or field.

---

## Example Request & Response

### Request

```http
GET /v1/books?order=fake_column:asc
```

### Response

```json
{
    "error": {
        "status": 400,
        "title": "Bad Request",
        "message": "Column 'fake_column' does not exist in table 'books'",
        "details": ["invalid order column: fake_column"],
        "documentation": "https://conduit.untapped.tech/docs/errors/bad-request"
    }
}
```

---

[← Back to Error Index](index.md)
