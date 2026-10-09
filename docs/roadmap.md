# Roadmap

Current status and forward-looking capabilities for Conduit.

---

## Current Release (v1.0.0)

Conduit core functionality focuses on high performance, multi-format input/output, and database portability:

- [x] **Multi-Format Serialization Engine:** Full input and output support for JSON, XML, YAML, TOML, NDJSON, CSV, and CBOR.
- [x] **Unified Database Drivers:** Support for SQLite, PostgreSQL, MySQL/MariaDB, SQL Server, libSQL/Turso, ClickHouse, and Oracle, plus an in-memory test driver.
- [x] **Schema API:** Dynamic table creation (`POST /v1/schema/<table>`), deletion (`DELETE /v1/schema/<table>`), and catalog introspection (`GET /v1/schema`).
- [x] **Full CRUD REST API:** Uniform endpoints for `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`.
- [x] **Advanced Query System:**
  - Tokenized lexer and AST parser for SQL-like `where` filters (`=`, `!=`, `<`, `>`, `<=`, `>=`, `LIKE`, `IN`, `NOT IN`, `IS NULL`, `IS NOT NULL`, `AND`, `OR`, `NOT`).
  - Column sorting with `order=<column>[:asc|:desc]`.
  - Dialect-aware SQL generation and parameter binding.
- [x] **LRU Performance Caching:** Compiled WHERE clauses, parameterized SQL statements, query metadata, and DB auth token lookups cached via in-memory thread-safe LRU.
- [x] **OpenAPI 3.0 & Interactive Docs:**
  - Live OpenAPI 3.0 specification generated from database schema at `/v1/openapi.json`.
  - Interactive Scalar API documentation UI rendered at `/v1/docs`.
  - CLI flag `--export-openapi <path>` for CI/CD pipelines.
- [x] **Configurable Routing & Base Paths:** Support for custom base paths (e.g. `/v1/`, `/api/`, or `/`).
- [x] **Pluggable Authorization:** Environment variables, database-backed token storage with caching, and No-Op fallback.
- [x] **Multi-Format Config System:** Automatic detection of `config.json`, `config.yaml`, `config.yml`, `config.toml`, and `config.xml`.

---

## Planned Capabilities

The following features are planned for subsequent minor and major releases:

### Extended Database Support
- Snowflake driver
- DuckDB driver
- Spanner driver

### Advanced Querying & Transformations
- Support for relational `JOIN` queries and foreign key graph traversal
- Aggregations (`COUNT`, `SUM`, `AVG`, `MIN`, `MAX`, `GROUP BY`)
- Dynamic calculated virtual columns

### Security & Governance
- OAuth2 / OIDC token verification provider
- Role-Based Access Control (RBAC) per table and column
- Row-Level Security (RLS) policies

### Event-Driven Ecosystem
- Server-Sent Events (SSE) and WebSockets for real-time table change notifications
- Webhooks triggered on table mutations
- Change Data Capture (CDC) streaming integration
