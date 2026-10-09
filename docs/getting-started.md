# Getting Started

Get your first Conduit server running in minutes.

Conduit ships with sensible defaults: an embedded SQLite database driver and no authorization required out of the box (`No-Op Auth`). This makes it easy to experiment locally without complex configuration.

---

## 1. Clone and Build

Download the Conduit source code and build the server binary:

```bash
git clone https://github.com/untappedtech/conduit
cd conduit
go build -o conduit ./cmd/server
```

You now have a static `conduit` binary ready to execute.

---

## 2. Running Conduit

### Using the Compiled Binary

```bash
./conduit --config ./config.yaml
```

### Using `go run`

```bash
go run ./cmd/server --config ./config.yaml
```

If no configuration file path is passed explicitly, Conduit automatically searches the working directory in the following order:

1. `config.json`
2. `config.yaml`
3. `config.yml`
4. `config.toml`
5. `config.xml`

---

## 3. Command-Line Flags

| Flag                         | Short | Description                                                                                                       |
| ---------------------------- | ----- | ----------------------------------------------------------------------------------------------------------------- |
| `--config <path>`            | `-c`  | Specifies the path to the configuration file (JSON, YAML, TOML, or XML).                                          |
| `--generate-config <format>` | `-g`  | Generates a starter configuration file of the specified type (`json`, `yaml`, `toml`, `xml`) and exits.           |
| `--export-openapi <path>`    | `-e`  | Exports the full OpenAPI 3.0 JSON specification derived from the database schema to the specified file and exits. |

### Generate a Starter Configuration

```bash
./conduit --generate-config yaml
```

This creates a `config.yaml` file with sensible defaults including server settings, SQLite database path, and policy rules.

### Export OpenAPI 3.0 Specification

```bash
./conduit --config ./config.yaml --export-openapi ./openapi.json
```

---

## 4. Your First Requests

With the server running (defaulting to `http://localhost:8080/v1`):

### Introspect Schema

Query the schema endpoint to inspect existing tables:

```bash
curl http://localhost:8080/v1/schema
```

If your database is empty, create your first table via the Schema API:

```bash
curl -X POST http://localhost:8080/v1/schema/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "columns": [
      { "name": "id", "type": "integer", "pk": true, "autoincrement": true },
      { "name": "title", "type": "text", "nullable": false },
      { "name": "completed", "type": "boolean", "default": "false" }
    ]
  }'
```

### Insert Data

```bash
curl -X POST http://localhost:8080/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Try Conduit", "completed": false}'
```

### Fetch Data in Multiple Formats

```bash
# Fetch as JSON (default)
curl http://localhost:8080/v1/tasks

# Fetch as CSV
curl http://localhost:8080/v1/tasks?format=csv

# Fetch as YAML
curl http://localhost:8080/v1/tasks?format=yaml
```

### Interactive API Documentation

Conduit automatically provides interactive API documentation powered by Scalar and OpenAPI 3.0 at:

- **Interactive UI**: `http://localhost:8080/v1/docs` (or `http://localhost:8080/docs`)
- **OpenAPI 3.0 JSON Spec**: `http://localhost:8080/v1/openapi.json` (or `http://localhost:8080/openapi.json`)

---

## Next Steps

- Explore [Configuration Overview](configuration.md) to customize server ports, database backends, and access policies.
- Check [CRUD API Reference](crud-api.md) for complete endpoint operations.
- Learn about advanced filtering and sorting in [Query API](query-api.md).
