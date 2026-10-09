# Pagination

Conduit uses offset-based pagination for all list endpoints (`GET /v1/<table>`). Pagination works across every table and every supported output format.

---

## Query Parameters

| Parameter | Type    | Default                          | Description                                                          |
| --------- | ------- | -------------------------------- | -------------------------------------------------------------------- |
| `limit`   | integer | `50` (or `server.default_limit`) | Maximum number of rows to return. Pass `limit=0` for unlimited mode. |
| `offset`  | integer | `0`                              | Zero-based index of records to skip before collecting rows.          |

---

## Basic Pagination

### First Page (Default)

Request the first 20 records:

```bash
GET /v1/books?limit=20
```

### Subsequent Pages

Use `offset` to page through larger datasets:

```bash
# Page 2 (rows 21-40)
GET /v1/books?limit=20&offset=20

# Page 3 (rows 41-60)
GET /v1/books?limit=20&offset=40
```

---

## Unlimited Mode (`limit=0`)

When clients need to stream or dump an entire dataset in a single request, pass `limit=0`:

```bash
GET /v1/books?limit=0&format=csv
```

This bypasses default limit constraints and returns all records matching any applied `where` filters.

---

## Combining Pagination, Sorting, and Filtering

Pagination executes cleanly alongside sorting (`order`) and filtering (`where`). The pipeline order is:

1. Filter rows matching the `where` expression.
2. Sort matching rows according to `order`.
3. Apply `offset` to skip initial rows.
4. Apply `limit` to cap result count.

```bash
GET /v1/customers?where=active = true&order=created_at:desc&limit=25&offset=50
```

---

## Multi-Format Support

Pagination operates identically regardless of format:

- **JSON / YAML / TOML / XML / CBOR:** The resulting array contains the sliced records.
- **CSV:** The CSV header row is written first, followed by the paginated row slice.
- **NDJSON:** Streams exactly the requested row slice, outputting one record per line.
