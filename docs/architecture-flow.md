# High-Level Architecture Flow

Every HTTP request that enters Conduit follows a predictable, deterministic execution lifecycle. Each stage in the pipeline performs a discrete validation or transformation step before handing off to the next layer.

---

## Request Pipeline Flowchart

```mermaid
sequenceDiagram
    autonumber
    actor Client as HTTP Client
    participant Dec as Format Decoder
    participant Hdr as HTTP Router
    participant Auth as Auth Chain
    participant Svc as API Service
    participant Cache as LRU Caches
    participant DB as DB Driver
    participant Enc as Format Encoder

    Client->>Dec: HTTP Request (Method, Path, Headers, Body)
    Dec->>Dec: Detect Content-Type & Decode Payload
    Dec->>Hdr: Parsed Record / Query
    Hdr->>Auth: Extract Token & Verify Role (readOnly / readWrite / admin)
    alt Unauthorized
        Auth-->>Client: 401 Unauthorized / 403 Forbidden
    end
    Auth->>Svc: Execute Operation Context
    alt Schema / Query Cached?
        Svc->>Cache: Lookup Schema & Compiled WHERE AST
        Cache-->>Svc: Hit
    else Cache Miss
        Svc->>DB: Introspect Schema / Parse WHERE AST
        DB-->>Svc: Column Metadata / New AST
        Svc->>Cache: Store in LRU Cache
    end
    Svc->>Svc: Validate Input Payload against Schema
    Svc->>DB: Execute Parameterized SQL Query
    DB-->>Svc: Query Results / Inserted Record
    Svc->>Enc: Format Result (JSON, XML, YAML, TOML, NDJSON, CSV, CBOR)
    Enc-->>Client: HTTP Response (Status, Content-Type, Serialized Body)
```

---

## Detailed Step-by-Step Breakdown

### 1. Request Ingestion & Decoding

When a request arrives, the **Format Decoder** inspects the `Content-Type` header:

- If body contains JSON, YAML, TOML, XML, CSV, NDJSON, or CBOR, it is parsed into normalized Go types (`map[string]any` or `[]map[string]any`).
- Query parameters (`where`, `order`, `limit`, `offset`, `format`) are extracted from the URL.

### 2. Route Matching

The **HTTP Router** checks the request against configured base paths:

- `/v1/schema` → Schema management routes (`handleSchema`).
- `/v1/docs` & `/v1/openapi.json` → Documentation and OpenAPI specs.
- `/v1/<table>` & `/v1/<table>/<id>` → Data routes (`handleCRUD`).

### 3. Authentication & Authorization

The **Authorization Chain** runs:

- Extracts token from `Authorization: Bearer <token>`, `Authorization: <token>`, `X-API-Key`, or `?token=`.
- Matches token against Environment variables or queries the Database Auth table (using LRU caching).
- Compares the token's granted role against the required operation permission (`readOnly`, `readWrite`, or `admin`).
- If unauthenticated, evaluates global `policy` flags (`publicReads`, `publicWrites`, `publicMutation`).

### 4. Service Logic & Caching

The **API Service** processes the operation:

- Checks the in-memory **Schema Cache** for column metadata. If missing, queries the driver and caches the result.
- For list queries with `where` clauses, checks the **AST Cache**. If missing, the lexer and parser generate a validated AST and cache it.
- Validates that fields exist, types match, and non-nullable fields are supplied.

### 5. Database Driver Execution

The **Database Driver** generates SQL tailored to the active database engine:

- Quotes table and column identifiers appropriately (e.g. `"id"` for SQLite/PostgreSQL, `` `id` `` for MySQL, `[id]` for SQL Server).
- Applies dialect-specific parameter placeholders (`?`, `$1`, `:p1`).
- Executes the query using connection pooling and returns row results.

### 6. Response Encoding

The **Format Encoder** converts results into the final payload:

- Uses `?format=` if provided; otherwise checks `Accept`, then echoes the input format, or falls back to JSON.
- Sorts fields according to table column schema order.
- Sets appropriate `Content-Type` headers and HTTP status codes (`200 OK`, `201 Created`, `204 No Content`).
