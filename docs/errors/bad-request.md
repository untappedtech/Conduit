# 400 Bad Request

The request could not be processed due to invalid parameters or malformed request syntax.

---

## What This Error Means

Conduit encountered a condition in the request preventing execution before touching the database. This frequently occurs when query parameters, column names, or payload structures fail syntactic validation.

---

## Common Causes

- Unknown column specified in `?order=` (e.g. `?order=non_existent_column:asc`).
- Unknown column referenced in `?where=` filter expressions.
- Query complexity violation: nested parenthetical grouping in `?where=` exceeding the configured `max_query_depth` threshold (default `5`).
- Missing required query parameters or malformed parameter values.
- Supplying invalid data types that cannot be parsed.

---

## How to Fix It

1. **Verify Column Names:** Ensure that columns referenced in `where` and `order` parameters match the schema of the target table. Check `GET /v1/schema/<table>` for the exact column names.
2. **Review Query Syntax:** Verify sorting direction syntax (`column` or `column:asc` or `column:desc`).
3. **Simplify Nested Filters:** If `max_query_depth` is exceeded, flatten nested parenthetical filter expressions or increase `max_query_depth` in `config.yaml`.
4. **Inspect Server Response Details:** Check the error message in the response for the specific offending parameter or field.

---

## Example Request & Response

### Unknown Column Example

```http
GET /v1/data/books?order=fake_column:asc
```

```json
{
  "error": "failed to validate query: column 'fake_column' does not exist in table 'books'",
  "code": 400
}
```

### Query Complexity Limit Exceeded Example

```http
GET /v1/data/books?where=((((((id > 10))))))
```

```json
{
  "error": "query too complex: nesting depth 6 exceeds maximum allowed depth of 5",
  "code": 400
}
```

---

[← Back to Error Index](index.md)
