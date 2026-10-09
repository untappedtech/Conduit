# CRUD API Reference

Conduit instantly maps every database table to standard REST endpoints. There are no hand-coded models, routes, or controllers needed.

All endpoints support JSON, XML, YAML, TOML, NDJSON, CSV, and CBOR.

---

## Endpoint Summary

| Method   | Endpoint           | Description                                              |
| -------- | ------------------ | -------------------------------------------------------- |
| `GET`    | `/v1/<table>`      | List records with filtering, sorting, and pagination.    |
| `GET`    | `/v1/<table>/<id>` | Fetch a single record by its primary key.                |
| `POST`   | `/v1/<table>`      | Insert a new record (or batch insert).                   |
| `PUT`    | `/v1/<table>/<id>` | Replace an existing record entirely.                     |
| `PATCH`  | `/v1/<table>/<id>` | Partially update specified fields on an existing record. |
| `DELETE` | `/v1/<table>/<id>` | Delete a record by its primary key.                      |

---

## 1. List Records (`GET /v1/<table>`)

Retrieves multiple records matching query criteria.

```bash
curl http://localhost:8080/v1/players
```

### Supported Query Parameters

| Parameter | Type    | Default | Description                                                                             |
| --------- | ------- | ------- | --------------------------------------------------------------------------------------- |
| `limit`   | integer | `50`    | Maximum number of rows to return. Pass `limit=0` for unlimited mode.                    |
| `offset`  | integer | `0`     | Number of rows to skip before returning results.                                        |
| `order`   | string  | none    | Sort expression: `<column>`, `<column>:asc`, or `<column>:desc`.                        |
| `where`   | string  | none    | SQL-like boolean expression: e.g. `age >= 21 AND active = true`.                        |
| `format`  | string  | `json`  | Response serialization format (`json`, `xml`, `yaml`, `toml`, `ndjson`, `csv`, `cbor`). |

### Complex Query Example

```bash
curl "http://localhost:8080/v1/players?where=score%20%3E%3D%2050&order=score:desc&limit=10&offset=0&format=yaml"
```

---

## 2. Retrieve Single Record (`GET /v1/<table>/<id>`)

Fetches a specific record identified by its primary key value.

```bash
curl http://localhost:8080/v1/players/42
```

### Response Example (JSON)

```json
{
    "id": 42,
    "name": "Wayne Gretzky",
    "number": 99,
    "team_id": 7
}
```

If no record exists matching the primary key, Conduit returns `404 Not Found`.

---

## 3. Insert Record (`POST /v1/<table>`)

Inserts one or more records into the table.

### Single Record Insert

```bash
curl -X POST http://localhost:8080/v1/players \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Connor McDavid",
    "number": 97,
    "team_id": 7
  }'
```

### Batch Ingestion

Conduit supports inserting an array or stream of records:

```bash
curl -X POST http://localhost:8080/v1/players \
  -H "Content-Type: application/json" \
  -d '[
    {"name": "Sidney Crosby", "number": 87, "team_id": 4},
    {"name": "Nathan MacKinnon", "number": 29, "team_id": 8}
  ]'
```

Returns HTTP `201 Created` with the inserted record(s), including newly assigned primary keys.

---

## 4. Replace Record (`PUT /v1/<table>/<id>`)

Replaces the entire record at `<id>`. All non-nullable columns must be provided in the payload.

```bash
curl -X PUT http://localhost:8080/v1/players/42 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wayne Gretzky",
    "number": 99,
    "team_id": 12
  }'
```

Returns HTTP `200 OK` with the updated record.

---

## 5. Partial Update (`PATCH /v1/<table>/<id>`)

Updates only the columns included in the request body, leaving untouched fields unmodified.

```bash
curl -X PATCH http://localhost:8080/v1/players/42 \
  -H "Content-Type: application/json" \
  -d '{
    "team_id": 15
  }'
```

Returns HTTP `200 OK` with the updated record.

---

## 6. Delete Record (`DELETE /v1/<table>/<id>`)

Deletes the record identified by `<id>`.

```bash
curl -X DELETE http://localhost:8080/v1/players/42
```

Returns HTTP `204 No Content` upon successful deletion. If the record does not exist, returns `404 Not Found`.
