# Formats Overview

Conduit is built around a versatile multi-format serialization engine. Every endpoint can consume and produce data across 7 different serialization formats without requiring custom controllers, serializers, or duplicate routes.

Clients choose their preferred format using standard HTTP headers (`Content-Type`, `Accept`) or the query parameter `?format=`.

---

## Supported Formats

| Format | MIME Type / Content-Type | Input | Output | Details |
| ------ | ------------------------ | :---: | :----: | ------- |
| **JSON** | `application/json` | ✅ | ✅ | Default format. Standard JSON objects and arrays. See [JSON Format](formats-json.md). |
| **XML** | `application/xml`, `text/xml` | ✅ | ✅ | Structured XML element tree with schema-ordered fields. See [XML Format](formats-xml.md). |
| **YAML** | `application/x-yaml`, `text/yaml` | ✅ | ✅ | Indentation-based document format. See [YAML Format](formats-yaml.md). |
| **TOML** | `application/toml`, `text/toml` | ✅ | ✅ | Structured table format using `[table]` / `[[table]]`. See [TOML Format](formats-toml.md). |
| **NDJSON** | `application/x-ndjson` | ✅ | ✅ | Newline-delimited JSON objects. Ideal for streaming and bulk loads. See [NDJSON Format](formats-ndjson.md). |
| **CSV** | `text/csv` | ✅ | ✅ | Comma-separated values with header rows. See [CSV Format](formats-csv.md). |
| **CBOR** | `application/cbor` | ✅ | ✅ | High-performance canonical binary serialization. See [CBOR Format](formats-cbor.md). |

---

## Format Negotiation Rules

When an incoming request is processed, Conduit determines the output format using a deterministic priority hierarchy:

```
1. ?format= query parameter (Explicit Override)
       ↓
2. Accept header (Content Negotiation)
       ↓
3. Content-Type of request body (Echo input format)
       ↓
4. Default fallback: JSON
```

### 1. Query Parameter Override (`?format=`)
The `?format=` parameter takes highest precedence:
- `?format=json`
- `?format=xml`
- `?format=yaml` (or `?format=yml`)
- `?format=toml`
- `?format=ndjson`
- `?format=csv`
- `?format=cbor`

```bash
curl http://localhost:8080/v1/players?format=toml
```

### 2. HTTP `Accept` Header
If no query parameter is provided, Conduit matches the `Accept` header:

```bash
curl -H "Accept: application/x-yaml" http://localhost:8080/v1/players
```

### 3. Echoing Request Body Format
If no output preference is expressed via query parameter or `Accept`, Conduit inspects the `Content-Type` header of the incoming payload. If you submit a `POST` or `PUT` request with `Content-Type: application/xml`, the newly created or updated record is returned as XML.

### 4. Default Fallback
If no preferences are specified, Conduit defaults to standard `application/json`.

---

## Input Payloads

Conduit decodes incoming payloads into typed Go records and validates them against the active table schema:

```bash
# Insert via JSON
curl -X POST http://localhost:8080/v1/players \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "score": 100}'

# Insert via YAML
curl -X POST http://localhost:8080/v1/players \
  -H "Content-Type: application/x-yaml" \
  -d '
name: Bob
score: 85
'

# Insert via CSV
curl -X POST http://localhost:8080/v1/players \
  -H "Content-Type: text/csv" \
  -d 'name,score
Charlie,92'
```

---

## Output Consistency

All schema-aware formats (JSON, XML, YAML, TOML, NDJSON, CSV) preserve column ordering defined in the database schema. Any dynamic or extra fields returned are consistently sorted alphabetically.
