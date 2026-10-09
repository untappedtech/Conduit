# OpenAPI Configuration

The `openapi` configuration block customizes the OpenAPI 3.0.3 specification and interactive documentation UI served automatically by Conduit.

Conduit dynamically generates an OpenAPI specification by inspecting the active relational database schema. You can customize the API title, version, description, terms of service, contact info, license details, and server endpoints.

---

## Configuration Settings

```yaml
openapi:
    title: "Conduit API"
    version: "1.0.0"
    description: "Multi-Format, Database-Agnostic REST Engine API documentation automatically generated from relational database schema."
    terms_of_service: "https://example.com/terms"
    contact:
        name: "API Support Team"
        email: "support@example.com"
        url: "https://example.com/support"
    license:
        name: "MIT"
        url: "https://opensource.org/licenses/MIT"
    servers:
        - url: "http://localhost:8080/v1"
          description: "Local development server"
        - url: "https://api.example.com/v1"
          description: "Production gateway"
```

### Settings Reference

| Field                   | Type     | Default                                                | Description                                                                      |
| ----------------------- | -------- | ------------------------------------------------------ | -------------------------------------------------------------------------------- |
| `title`                 | `string` | `"Conduit API"`                                        | Title displayed in the OpenAPI specification and interactive documentation UI.   |
| `version`               | `string` | `"1.0.0"`                                              | API version string in the OpenAPI `info` block.                                  |
| `description`           | `string` | `"Multi-Format, Database-Agnostic REST Engine API..."` | Description text appearing in the API documentation header.                      |
| `terms_of_service`      | `string` | `""`                                                   | URL to the terms of service governing API access.                                |
| `contact`               | `object` | `null`                                                 | Contact information for the API maintenance team.                                |
| `contact.name`          | `string` | `""`                                                   | Identifying name of the contact person or organization.                          |
| `contact.email`         | `string` | `""`                                                   | Support contact email address.                                                   |
| `contact.url`           | `string` | `""`                                                   | Support or project website URL.                                                  |
| `license`               | `object` | `null`                                                 | Licensing information applied to the exposed API.                                |
| `license.name`          | `string` | `""`                                                   | License identifier name (e.g., `MIT`, `Apache 2.0`).                             |
| `license.url`           | `string` | `""`                                                   | Link to the full license text.                                                   |
| `servers`               | `array`  | `[]`                                                   | List of target servers where the API is hosted.                                  |
| `servers[].url`         | `string` | `""`                                                   | Target server base URL. If omitted, Conduit derives the active listener address. |
| `servers[].description` | `string` | `""`                                                   | Human-readable description of the target server environment.                     |

---

## Interactive Documentation UI

Conduit includes an embedded interactive API reference powered by [Scalar](https://scalar.com).

Whenever Conduit is running, the interactive documentation is accessible in your web browser at:

```
http://localhost:8080/v1/docs
```

Root fallback aliases are also registered at `http://localhost:8080/docs`.

The browser interface renders an interactive sandbox where developers can browse all discovered database tables, inspect columns and data types, test query parameters, and execute live API requests.

---

## OpenAPI Spec Endpoint

The raw OpenAPI 3.0.3 specification is served at:

```
GET /v1/openapi.json
```

A root alias is also available at `GET /openapi.json`.

Clients, API gateways, SDK generators, and documentation portals can fetch the live JSON spec at any time:

```bash
curl http://localhost:8080/v1/openapi.json
```

The endpoint responds with `Content-Type: application/json; charset=utf-8` and includes CORS headers (`Access-Control-Allow-Origin: *`) to enable browser-based API tooling out of the box.

---

## Live Schema Synchronization

The OpenAPI specification is generated directly from live database table metadata and cached in memory.

When table schemas change via the Schema API:

- `POST /v1/schema/<table>` (Create table)
- `DELETE /v1/schema/<table>` (Drop table)

Conduit automatically invalidates the cached OpenAPI document. The next request to `/v1/openapi.json` or `/v1/docs` dynamically rebuilds the specification with the updated tables, column types, and CRUD endpoints without requiring a server restart.

---

## Command-Line Export

You can export the generated OpenAPI specification to a static file without starting the long-running HTTP server using the `-e` / `--export-openapi` CLI flag:

```bash
conduit --config config.yaml --export-openapi ./openapi.json
```

This is ideal for CI/CD pipelines, documentation site generation, or automated client SDK builds.
