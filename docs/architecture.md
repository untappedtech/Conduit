# Architecture Overview

Conduit is designed around clean-architecture principles with strictly separated, modular components. Each layer has a single responsibility, ensuring the engine remains predictable, high-performing, testable, and straightforward to extend.

---

## Architectural Diagram

```
                    ┌──────────────────────────────────────────┐
                    │               HTTP Client                │
                    │   (cURL, Browser, Microservice, ETL)     │
                    └────────────────────┬─────────────────────┘
                                         │ HTTP Request
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │          Format Decoder Engine           │
                    │    JSON / XML / YAML / TOML / CSV...     │
                    └────────────────────┬─────────────────────┘
                                         │ Decoded Record / Query
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │            HTTP Server & Mux             │
                    │        (Routing & Path Matching)         │
                    └────────────────────┬─────────────────────┘
                                         │
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │           Authorization Chain            │
                    │      (Env Auth, DB Auth, No-Op)          │
                    └────────────────────┬─────────────────────┘
                                         │ Authenticated Context
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │             API Service Core             │
                    │   - Schema Caching                       │
                    │   - Query WHERE Lexer & AST Caching      │
                    │   - Schema-Based Field Validation        │
                    └────────────────────┬─────────────────────┘
                                         │ Validated SQL Operation
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │          Database Driver Layer           │
                    │   SQLite, Postgres, MySQL, SQL Server,   │
                    │   libSQL, ClickHouse, Oracle, Memory     │
                    └────────────────────┬─────────────────────┘
                                         │ Relational SQL
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │            SQL Database Engine           │
                    └────────────────────┬─────────────────────┘
                                         │ Row Data
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │          Format Encoder Engine           │
                    │   JSON / XML / YAML / TOML / NDJSON...   │
                    └────────────────────┬─────────────────────┘
                                         │ HTTP Response
                                         ▼
                    ┌──────────────────────────────────────────┐
                    │               HTTP Client                │
                    └──────────────────────────────────────────┘
```

---

## Component Responsibilities

### 1. Format Decoder Engine (`internal/http/decoder.go`)

Inspects incoming `Content-Type` headers and converts raw HTTP request bodies into structured Go maps or record arrays. Supports JSON, XML, YAML, TOML, NDJSON, CSV, and CBOR.

### 2. Router & HTTP Handler (`internal/http/handler.go`)

Provides clean path mapping:

- `/v1/schema` and `/v1/schema/<table>` for schema introspection and DDL management.
- `/v1/<table>` and `/v1/<table>/<id>` for CRUD data manipulation.
- `/v1/openapi.json` and `/v1/docs` for OpenAPI specs and interactive documentation UI.

### 3. Authorization Chain (`internal/auth`)

Evaluates request permissions before database queries are made. Implements token extraction from headers and query parameters, and supports environment variable tokens and database-backed tokens with in-memory LRU caching.

### 4. API Service Layer (`internal/service`)

Coordinates application logic:

- Caches table schemas to eliminate redundant introspection queries.
- Parses SQL-like `where` filters via a lexer and AST parser, reusing compiled ASTs via an LRU cache.
- Validates data payloads against schema column types and constraints.

### 5. Database Drivers (`internal/db`)

Unified interface providing:

- Dialect-aware SQL builders (quoting identifiers, formatting placeholders).
- Schema introspection (`ListTables`, `GetSchema`).
- Table creation and deletion (`CreateTable`, `DropTable`).
- Parameterized CRUD queries (`Insert`, `Get`, `Update`, `Patch`, `Delete`, `List`).

### 6. Format Encoder Engine (`internal/http/encoder.go`)

Formats data records into the requested output format (`?format=` parameter or `Accept` header), respecting database column ordering and emitting standard content headers.

### 7. Structured Error System (`internal/errors`)

Translates application failures into standardized RFC-compliant error responses with machine-readable status codes, localized messages, and documentation links.

---

## Further Reading

- [High-Level Flow](architecture-flow.md) — Detailed trace of a request through the system.
- [Multi-Format Engine](architecture-multi-format-engine.md) — Deep dive into serialization and content negotiation.
- [Database Drivers](architecture-database-drivers.md) — Driver architecture and dialect handling.
- [Error System](architecture-error-system.md) — Structured error catalogs and response envelopes.
