# OpenAPI & Documentation UI

Conduit automatically inspects your active relational database schema and produces an **OpenAPI 3.0.3 specification** and an embedded interactive API documentation browser powered by **Scalar**.

There is no manual schema drafting or synchronization required. When database tables are created or dropped, the documentation and OpenAPI specification update in real time.

---

## Interactive Documentation Browser

Conduit serves the interactive documentation user interface at:

```
GET /v1/docs
```

> **Note:** Conduit also mounts a root alias at `GET /docs` for convenience.

Visiting this URL in any modern web browser opens the Scalar API reference interface, providing:

- Full endpoint catalog for all database tables and schema management operations.
- Interactive request runner to test queries, filters, and mutations live against the database.
- Schema definitions showing table columns, SQL data types, nullability, defaults, and primary keys.
- Example payloads across all supported formats (JSON, XML, YAML, TOML, NDJSON, CSV, CBOR).

---

## OpenAPI Specification Endpoint

The machine-readable OpenAPI 3.0.3 JSON document is available at:

```
GET /v1/openapi.json
```

Root alias: `GET /openapi.json`.

```bash
curl http://localhost:8080/v1/openapi.json
```

### Response Headers

- `Content-Type: application/json; charset=utf-8`
- `Access-Control-Allow-Origin: *` (CORS enabled for web tools like Swagger Editor, Postman, and Insomnia)
- `Cache-Control: no-cache`

---

## Configuration

You can customize the OpenAPI specification and documentation UI using the `openapi` section in your configuration file:

```yaml
openapi:
    title: "Inventory Service API"
    version: "2.1.0"
    description: "Production inventory and warehouse management REST engine."
    terms_of_service: "https://example.com/terms"
    contact:
        name: "DevOps Team"
        email: "devops@example.com"
        url: "https://example.com/team"
    license:
        name: "Apache 2.0"
        url: "https://www.apache.org/licenses/LICENSE-2.0"
    servers:
        - url: "https://api.example.com/v1"
          description: "Production Gateway"
        - url: "http://localhost:8080/v1"
          description: "Local Development"
```

For detailed field descriptions, see [OpenAPI Configuration](config-openapi.md).

---

## Static CLI Export

In addition to serving the spec over HTTP, Conduit can generate and export the specification directly to a file from the command line without keeping the HTTP server running:

```bash
conduit --config ./config.yaml --export-openapi ./openapi.json
# Or using short flags:
conduit -c ./config.yaml -e ./openapi.json
```

This makes it easy to integrate Conduit into automated workflows, such as:

- Generating client SDKs using OpenAPI Generator or Stainless.
- Publishing API contracts to developer portals and catalogs.
- Running contract and schema validation in CI/CD pipelines.

---

## Dynamic Cache Invalidation

The OpenAPI engine caches the generated JSON document in memory for maximum request throughput.

When table structures change dynamically via the [Schema API](schema-api.md):

- `POST /v1/schema/<table>` (New table created)
- `DELETE /v1/schema/<table>` (Table dropped)

Conduit immediately clears the internal OpenAPI cache. Subsequent requests to `/v1/openapi.json` and `/v1/docs` regenerate the specification with zero downtime.
