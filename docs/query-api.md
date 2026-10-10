# Query API Reference

Conduit features a SQL-like query filtering engine via `where` and column sorting via `order` across all table data endpoints (`GET /v1/data/<table>`).

Expressions are tokenized by an internal lexer, parsed into an Abstract Syntax Tree (AST), validated against table columns, and compiled into parameterized SQL using dialect-specific placeholder syntax (`?` for SQLite/MySQL, `$1` for PostgreSQL, `:p1` for Oracle, etc.).

---

## Query Parameters

| Parameter | Type      | Default | Description                                                                           |
| --------- | --------- | ------- | ------------------------------------------------------------------------------------- |
| `where`   | `string`  | none    | SQL-like boolean filtering expression compiled to parameterized SQL.                  |
| `order`   | `string`  | none    | Column sort expression: `<column>`, `<column>:asc`, or `<column>:desc`.               |
| `limit`   | `integer` | `50`    | Maximum rows to return. Pass `limit=0` for unlimited mode.                            |
| `offset`  | `integer` | `0`     | Number of rows to skip before returning results.                                      |
| `format`  | `string`  | `json`  | Output serialization format (`json`, `xml`, `yaml`, `toml`, `ndjson`, `csv`, `cbor`). |

---

## Sorting with `order`

The `order` parameter specifies the column to sort by and the direction. The column must exist on the table.

```bash
# Ascending sort (default)
GET /v1/data/books?order=title
GET /v1/data/books?order=published_year:asc

# Descending sort
GET /v1/data/books?order=published_year:desc
```

If an unknown column name is supplied, Conduit returns `400 Bad Request`.

---

## Filtering with `where`

The `where` parameter accepts SQL-like boolean expressions with literals, comparisons, set inclusion, and grouping parentheses:

```bash
GET /v1/data/books?where=author = 'Asimov'
GET /v1/data/books?where=published_year >= 1950 AND published_year <= 1980
GET /v1/data/books?where=title LIKE '%Foundation%'
GET /v1/data/books?where=status IN ('published', 'draft')
GET /v1/data/books?where=deleted_at IS NULL
GET /v1/data/books?where=(rating >= 4.5 OR featured = true) AND archived = false
```

---

## Supported Operators

| Operator      | Meaning                                                 | Example                              |
| ------------- | ------------------------------------------------------- | ------------------------------------ |
| `=`, `==`     | Equality                                                | `status = 'active'`                  |
| `!=`, `<>`    | Inequality                                              | `status != 'archived'`               |
| `<`, `<=`     | Less than / Less than or equal                          | `price <= 29.99`                     |
| `>`, `>=`     | Greater than / Greater than or equal                    | `score >= 100`                       |
| `LIKE`        | String pattern matching (`%` wildcard, `_` single char) | `name LIKE 'Alice%'`                 |
| `IN`          | Set inclusion list                                      | `category IN ('tech', 'science')`    |
| `NOT IN`      | Negative set inclusion                                  | `role NOT IN ('guest', 'banned')`    |
| `IS NULL`     | Null verification                                       | `deleted_at IS NULL`                 |
| `IS NOT NULL` | Non-null verification                                   | `verified_at IS NOT NULL`            |
| `AND`         | Logical conjunction                                     | `age >= 18 AND status = 'active'`    |
| `OR`          | Logical disjunction                                     | `tier = 'gold' OR tier = 'platinum'` |
| `NOT`         | Logical negation                                        | `NOT (status = 'banned')`            |

---

## Data Types & Literals

- **Strings:** Single or double quotes (e.g. `'hello'`, `"world"`). Embedded quotes: `'It\'s'` or `'It''s'`.
- **Numbers:** Integers and floating-point literals (e.g. `42`, `-15`, `3.1415`).
- **Booleans:** `true` and `false` (case-insensitive).
- **Null:** `NULL` (case-insensitive).

---

## Query Complexity Bounding

To protect against deeply recursive expressions and computational exhaustion attacks, Conduit enforces a strict **AST Complexity Bouncer**:

- Nesting depth is tracked whenever opening parentheses `(` are encountered in sub-expressions.
- The maximum allowed nesting depth is configured via `max_query_depth` in your server configuration (default: `5`).
- Sibling groupings (such as `(a = 1) AND (b = 2)`) only reach depth 1 and are fully permitted.
- If nested expressions exceed `max_query_depth`, parsing immediately halts and Conduit returns `400 Bad Request`:

```json
{
    "error": {
        "status": 400,
        "title": "Bad Request",
        "message": "query too complex: nesting depth 6 exceeds maximum allowed depth of 5",
        "documentation_url": "https://conduit.untapped.tech/docs/errors/bad-request"
    }
}
```

---

## Performance & AST Caching

Conduit includes a thread-safe LRU cache for compiled WHERE expressions:

- Incoming `where` strings are parsed into AST trees and cached alongside the target table schema and depth bounds.
- Subsequent identical expressions bypass lexing and parsing, jumping directly to parameterized SQL execution.
- Dialect placeholders are safely bound to prevent SQL injection vulnerabilities.

---

## Error Responses

- **`400 Bad Request`:** Filtered or sorted column does not exist on the target table, or query nesting depth exceeds `max_query_depth`.
- **`422 Unprocessable Entity`:** Malformed expression syntax, mismatched quotes, or unclosed parentheses.
- **`500 Internal Server Error`:** Database driver execution failure.
