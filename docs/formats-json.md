# JSON Format

JSON is Conduit's default data format. If no `Content-Type`, `Accept`, or `?format=` parameter is specified, Conduit automatically parses input as JSON and renders responses as JSON.

- **MIME Type:** `application/json`
- **Query Parameter:** `?format=json`
- **Supported Operations:** Input (Decode) & Output (Encode)

---

## Sending JSON Input

To send JSON data to Conduit, set the `Content-Type: application/json` header:

### Create Record (POST)

```bash
curl -X POST http://localhost:8080/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Dune",
    "author": "Frank Herbert",
    "published_year": 1965,
    "available": true
  }'
```

### Batch Ingestion (POST)

Conduit accepts an array of objects for bulk insertion:

```bash
curl -X POST http://localhost:8080/v1/books \
  -H "Content-Type: application/json" \
  -d '[
    { "title": "Neuromancer", "author": "William Gibson", "published_year": 1984 },
    { "title": "Snow Crash", "author": "Neal Stephenson", "published_year": 1992 }
  ]'
```

---

## Receiving JSON Output

JSON is the default response format. You can also explicitly request JSON:

```bash
curl http://localhost:8080/v1/books?format=json
```

Or using headers:

```bash
curl -H "Accept: application/json" http://localhost:8080/v1/books
```

### Single Record Response

```json
{
    "id": 1,
    "title": "Dune",
    "author": "Frank Herbert",
    "published_year": 1965,
    "available": true
}
```

### List Response

```json
[
    {
        "id": 1,
        "title": "Dune",
        "author": "Frank Herbert",
        "published_year": 1965,
        "available": true
    },
    {
        "id": 2,
        "title": "Neuromancer",
        "author": "William Gibson",
        "published_year": 1984,
        "available": true
    }
]
```

---

## Key Characteristics

- **Schema Ordering:** JSON keys in responses mirror the column ordering defined in the database table schema.
- **Strong Typing:** Numbers, booleans, and nulls maintain their native JSON primitive types. Date and timestamp values are serialized in ISO-8601 string format.
