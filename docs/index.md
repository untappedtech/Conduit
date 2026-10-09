# Conduit Documentation

Welcome to the Conduit documentation.

Conduit is a **zero-boilerplate REST engine** written in Go that instantly exposes your relational databases as a fully-typed, multi-format HTTP API supporting **JSON, XML, YAML, TOML, NDJSON, CSV, and CBOR** with pluggable database drivers and flexible authorization.

---

## Documentation Sections

### Core
- [Philosophy](philosophy.md) — The core design principles, mission, and reasons behind Conduit.
- [Getting Started](getting-started.md) — Clone, build, run, and issue your first API requests in minutes.
- [Roadmap](roadmap.md) — Current release status, planned features, and future directions.

### Formats
- [Formats Overview](formats.md) — How content negotiation, request parsing, and response encoding work.
- [JSON Format](formats-json.md) — Default lightweight JSON input and output.
- [XML Format](formats-xml.md) — Stable element tree serialization with XML request and response structures.
- [YAML Format](formats-yaml.md) — Human-readable indentation-based payloads.
- [TOML Format](formats-toml.md) — Configuration and table-oriented serialization.
- [NDJSON Format](formats-ndjson.md) — Newline-delimited JSON for high-throughput streaming and bulk ingestion.
- [CSV Format](formats-csv.md) — Tabular comma-separated values for spreadsheets and bulk ingestion.
- [CBOR Format](formats-cbor.md) — Compact binary object representation for efficient machine-to-machine exchange.

### Configuration
- [Configuration Overview](configuration.md) — Configuration file detection order, structure, and CLI options.
- [Server Settings](config-server.md) — Bind address, port, base paths, and pagination limits.
- [Database Settings](config-database.md) — Driver options, DSN connection strings, and driver examples.
- [Policy Settings](config-policy.md) — Global unauthenticated access permissions (`publicReads`, `publicWrites`, `publicMutation`).
- [Authorization Overview](config-auth.md) — Pluggable authentication strategies and token role mappings.
- [Environment Auth](config-auth-env.md) — Token authentication configured via environment variables.
- [Database Auth](config-auth-db.md) — Dynamic token lookup backed by an operational or auth database table with LRU caching.
- [No-Op Auth](config-auth-noop.md) — Permissive authentication mode for open and development setups.

### API Reference
- [Schema API](schema-api.md) — Inspect database tables, create new tables, and drop tables via HTTP.
- [CRUD API](crud-api.md) — Full REST operations (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) for all database tables.
- [Query API](query-api.md) — SQL-like filtering expressions (`where`), sorting (`order`), and parameter reference.
- [Pagination](pagination.md) — Offset-based pagination with `limit` and `offset` across all formats.
- [Validation](validation.md) — Database schema-driven payload validation, nullability, and type enforcement.

### Architecture
- [Architecture Overview](architecture.md) — Layered architecture, component responsibilities, and request lifecycle.
- [High-Level Flow](architecture-flow.md) — Step-by-step pipeline from HTTP request to database driver and response.
- [Multi-Format Engine](architecture-multi-format-engine.md) — Multi-format decoders, schema-aware encoders, and fallback semantics.
- [Database Drivers](architecture-database-drivers.md) — Pluggable driver interface, query dialects, and supported engines.
- [Error System](architecture-error-system.md) — RFC-style error envelopes, localized error catalog, and documentation links.

### Error Catalog
- [Error Index](errors/index.md) — Comprehensive guide to Conduit error codes and troubleshooting.
- [400 Bad Request](errors/bad-request.md)
- [401 Unauthorized](errors/unauthorized.md)
- [403 Forbidden](errors/forbidden.md)
- [404 Not Found](errors/not-found.md)
- [405 Method Not Allowed](errors/method-not-allowed.md)
- [409 Conflict](errors/conflict.md)
- [415 Unsupported Media Type](errors/unsupported-media-type.md)
- [422 Unprocessable Entity](errors/unprocessable-entity.md)
- [429 Too Many Requests](errors/too-many-requests.md)
- [500 Internal Server Error](errors/internal-server-error.md)
- [501 Not Implemented](errors/not-implemented.md)
- [503 Service Unavailable](errors/service-unavailable.md)
