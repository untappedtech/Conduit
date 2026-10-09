# 422 Unprocessable Entity

The request payload or query expression is syntactically well-formed, but contains semantic validation errors.

---

## What This Error Means

Conduit was able to parse the incoming request format (e.g. valid JSON or YAML), but the content violates validation rules, such as missing required `NOT NULL` fields, incompatible data types, or a malformed `where` query expression.

---

## Common Causes

- Missing a required `NOT NULL` column in `POST` or `PUT` payloads.
- Syntax errors in a `where` query string (such as unclosed parentheses, missing quotes, or unknown operators).
- Passing incompatible data types that cannot be coerced (e.g. sending a string for an integer column).
- Passing an empty payload or empty column list when creating a table.

---

## How to Fix It

1. **Verify Required Columns:** Check the schema (`GET /v1/schema/<table>`) for all non-nullable columns and supply them in your payload.
2. **Review `where` Expressions:** Check SQL expression syntax in `where` query parameters. Ensure parentheses are balanced and string literals are quoted.
3. **Inspect Error Details:** The `details` array will specify the exact column or token that caused the validation failure.

---

## Example Request & Response

### Request

```http
POST /v1/books
Content-Type: application/json

{
  "author": "Isaac Asimov"
}
```

### Response

```json
{
    "error": {
        "status": 422,
        "title": "Unprocessable Entity",
        "message": "Validation failed for table 'books'",
        "details": ["Column 'title' is required and cannot be null"],
        "documentation": "https://conduit.untapped.tech/docs/errors/unprocessable-entity"
    }
}
```

---

[← Back to Error Index](index.md)
