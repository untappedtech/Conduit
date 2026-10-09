# Philosophy

The principles and design tenets that shape Conduit.

> **Conduit's mission:** Eliminate boilerplate and expose relational databases as a clean, predictable, typed, multi-format HTTP API — without models, controllers, routing configurations, or custom serializers.

---

## Why Conduit Exists

Modern backend architectures frequently spend excessive engineering effort repeatedly solving the exact same problem: mapping relational database tables to standard HTTP REST endpoints.

Every new database table typically requires:

1. ORM / model definitions
2. Serializers and format encoders
3. Routing table registrations
4. Controller logic for standard CRUD operations
5. Filtering and sorting query mappers
6. Error envelope wrappers and documentation

**Conduit eliminates this layer entirely.** By connecting directly to your database, inspecting its schema, and dynamically establishing REST endpoints, Conduit delivers:

- **Instant Productivity:** Turn any new or existing SQL database into a web API immediately.
- **Protocol Flexibility:** Support clients consuming JSON, YAML, TOML, XML, CSV, NDJSON, and CBOR without building separate endpoints.
- **Database Portability:** Write standard queries against SQLite, PostgreSQL, MySQL/MariaDB, SQL Server, libSQL/Turso, ClickHouse, or Oracle using a uniform HTTP interface.

---

## Core Principles

### 1. Zero Boilerplate

You should not need to write Go code, define structs, or configure ORMs just to expose tables over HTTP. Point Conduit at a database DSN, launch the process, and every table is live.

### 2. Multi-Format as a First-Class Citizen

Data serialization is not an afterthought. Clients have different needs:

- Web browsers prefer **JSON**.
- Operations and configuration tooling rely on **YAML** and **TOML**.
- Legacy enterprise integrations require **XML**.
- Data analysts and spreadsheets work in **CSV**.
- Real-time pipelines stream **NDJSON**.
- IoT and low-bandwidth systems utilize binary **CBOR**.

Conduit supports all 7 formats natively on every endpoint, honoring HTTP content negotiation (`Content-Type` and `Accept`) as well as explicit `?format=` parameters.

### 3. Database Agnostic

The API surface remains identical whether your application is backed by SQLite, PostgreSQL, MySQL, MariaDB, SQL Server, libSQL (Turso), ClickHouse, or Oracle. The underlying driver translates filters, schema operations, and mutations to dialect-compliant SQL with secure parameterization.

### 4. Schema-Driven Truth

The database schema is the single source of truth:

- Fields, primary keys, nullability, and constraints are determined directly from the database catalog.
- Incoming payloads are strictly validated against schema definitions prior to executing SQL.
- When tables are added or modified via the Schema API (`POST /v1/schema/<table>`), endpoints and OpenAPI specifications update dynamically.

### 5. Predictable Error Handling

Errors should never leak internal stack traces or produce unpredictable formats. Conduit uses an RFC-style structured error envelope with machine-readable status codes, clear error titles, helpful details, and direct links to documentation.

### 6. Layered and Pluggable Design

Every component within Conduit has a dedicated responsibility:

- **Format Decoders & Encoders:** Handle request and response serialization.
- **Authorization Chains:** Evaluate permissions before database touches occur.
- **Service Layer:** Coordinates schema caching, query parsing, and validation.
- **Database Drivers:** Execute dialect-specific SQL safely.

---

## When to Use Conduit

- **Rapid Prototyping:** Build prototypes and frontends without writing backend boilerplate.
- **Internal Tools & Admin Dashboards:** Expose relational operational data cleanly to internal apps.
- **Data Ingestion & Extraction:** Ingest or export CSV, NDJSON, or JSON streams directly to and from SQL tables.
- **Multi-Format Integration Hubs:** Provide heterogeneous enterprise systems with the exact data format they need.
- **Database-First Applications:** Build applications where the SQL schema defines the architecture.
