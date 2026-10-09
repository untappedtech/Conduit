# Schema API Reference

The Schema API provides HTTP endpoints to inspect the database catalog, examine table column metadata, dynamically create new tables, and drop existing tables.

All schema endpoints support Conduit's multi-format engine (JSON, YAML, TOML, XML, CSV).

---

## Endpoints

| Method   | Endpoint             | Description                                         | Permission Required         |
| -------- | -------------------- | --------------------------------------------------- | --------------------------- |
| `GET`    | `/v1/schema`         | List all tables present in the database.            | `readOnly` or `publicReads` |
| `GET`    | `/v1/schema/<table>` | Introspect column definitions for a specific table. | `readOnly` or `publicReads` |
| `POST`   | `/v1/schema/<table>` | Create a new table with specified columns.          | `admin` or `publicMutation` |
| `DELETE` | `/v1/schema/<table>` | Drop a table from the database.                     | `admin` or `publicMutation` |

---

## 1. List Tables (`GET /v1/schema`)

Returns an array of all table names currently recognized by the database driver.

```bash
curl http://localhost:8080/v1/schema
```

### Response Example (JSON)

```json
["books", "authors", "categories"]
```

### Response Example (YAML)

```bash
curl http://localhost:8080/v1/schema?format=yaml
```

```yaml
- books
- authors
- categories
```

---

## 2. Inspect Table Schema (`GET /v1/schema/<table>`)

Returns detailed metadata about each column in the specified table.

```bash
curl http://localhost:8080/v1/schema/books
```

### Response Example

```json
[
    {
        "name": "id",
        "type": "integer",
        "pk": true,
        "nullable": false,
        "autoincrement": true
    },
    {
        "name": "title",
        "type": "text",
        "pk": false,
        "nullable": false
    },
    {
        "name": "published_year",
        "type": "integer",
        "pk": false,
        "nullable": true
    }
]
```

---

## 3. Create Table (`POST /v1/schema/<table>`)

Creates a new SQL table directly via HTTP. Conduit parses the payload and generates dialect-accurate DDL statements for the connected database engine.

### Request Body Format

```json
{
    "columns": [
        {
            "name": "id",
            "type": "integer",
            "pk": true,
            "autoincrement": true
        },
        {
            "name": "title",
            "type": "text",
            "nullable": false
        },
        {
            "name": "author_id",
            "type": "integer",
            "nullable": true
        },
        {
            "name": "price",
            "type": "real",
            "default": "0.0"
        },
        {
            "name": "in_stock",
            "type": "boolean",
            "default": "true"
        }
    ]
}
```

```bash
curl -X POST http://localhost:8080/v1/schema/inventory \
  -H "Content-Type: application/json" \
  -d '{
    "columns": [
      { "name": "sku", "type": "text", "pk": true },
      { "name": "quantity", "type": "integer", "nullable": false }
    ]
  }'
```

### Column Definition Attributes

| Attribute       | Type      | Description                                                                         |
| --------------- | --------- | ----------------------------------------------------------------------------------- |
| `name`          | `string`  | **Required.** Column identifier name.                                               |
| `type`          | `string`  | **Required.** SQL type (`integer`, `text`, `boolean`, `real`, `timestamp`, `blob`). |
| `pk`            | `boolean` | Set to `true` to declare this column as primary key.                                |
| `autoincrement` | `boolean` | Set to `true` for auto-incrementing integer identifiers.                            |
| `nullable`      | `boolean` | Whether column accepts `NULL`. Defaults to `true` unless `nullable: false`.         |
| `default`       | `string`  | Default value expression or literal.                                                |

> **Idempotency Note:** If the table already exists, Conduit does not throw an error; it returns the existing column definition schema with HTTP status `200 OK`.

---

## 4. Drop Table (`DELETE /v1/schema/<table>`)

Deletes an entire table and all its stored records from the database.

```bash
curl -X DELETE http://localhost:8080/v1/schema/inventory
```

### Response

Returns HTTP `204 No Content` upon successful deletion.

---

## Live OpenAPI Spec Invalidation

Whenever a table is created via `POST /v1/schema/<table>` or dropped via `DELETE /v1/schema/<table>`, Conduit immediately invalidates its internal OpenAPI generator cache. The `/v1/openapi.json` spec and `/v1/docs` interactive Scalar UI automatically reflect the new tables and endpoints on the very next request.
