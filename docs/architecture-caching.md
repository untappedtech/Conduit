# Caching Architecture

Conduit employs a multi-tiered, thread-safe in-memory caching system to maximize throughput and minimize latency on high-volume database workloads.

Because database schema introspection and query expression parsing are computationally expensive operations, Conduit caches them in memory using specialized Least Recently Used (LRU) data structures.

---

## Caching Tiers

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Incoming HTTP Request                           │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
    1. Auth Token Cache             ▼
   ┌─────────────────────────────────────────────────────────────────┐
   │ LRU Cache (Capacity & TTL) for DB Auth Provider Token Lookups   │
   └────────────────────────────────┬────────────────────────────────┘
                                    │
    2. Query AST & SQL Cache        ▼
   ┌─────────────────────────────────────────────────────────────────┐
   │ xxHash-keyed LRU Cache for WHERE Expression ASTs and Dialect SQL│
   └────────────────────────────────┬────────────────────────────────┘
                                    │
    3. Schema & Table Cache         ▼
   ┌─────────────────────────────────────────────────────────────────┐
   │ LRU Cache for Column Definitions & Table Catalog Metadata       │
   └────────────────────────────────┬────────────────────────────────┘
                                    │
                                    ▼
                         Database Engine Execution
```

---

## 1. Schema & Table Metadata Caching

Database drivers wrap their underlying connections in a `CachedDatabase` layer (`internal/db/cache.go`):

- **Table Catalog Caching:** The list of available database tables (`ListTables`) is cached in memory. Subsequent requests return the cached table list instantly.
- **Column Definition Caching:** The result of table column introspection (`Schema(tableName)`) is cached in a thread-safe LRU cache (default capacity: 256 tables).
- **Self-Invalidating Operations:** Whenever a table is created (`POST /v1/schema/<table>`) or dropped (`DELETE /v1/schema/<table>`), the schema cache is automatically invalidated. The cached tables list, schema entries, and compiled query caches are reset immediately.

---

## 2. Query Expression & SQL Caching

Complex SQL-like `where` filter expressions are processed in two stages (`internal/service/cache.go`):

### AST Cache

When a client submits a filter such as:

```
GET /v1/books?where=(rating >= 4.5 OR featured = true) AND archived = false
```

The expression string is combined with table column definitions and hashed using [xxHash](https://github.com/cespare/xxhash). If an AST already exists in the 1024-entry AST LRU cache, lexing and parsing are bypassed entirely.

### Dialect SQL Cache

The compiled parameterized SQL string and its argument templates are also cached by expression, column types, and database dialect. If the same query structure is requested again, Conduit reuses the compiled query structure directly, binding runtime arguments in constant time.

---

## 3. Database Auth Token Caching

When database-backed authentication is enabled (`auth.dbEnabled: true`), verifying API keys on every incoming request would result in an extra SQL query for every HTTP hit.

To eliminate this bottleneck, the database authenticator includes an in-memory token cache (`internal/auth/cache/lru_ttl.go`):

- **Configurable Capacity:** Number of distinct tokens retained in memory before least-recently-used eviction (configured via `auth.dbAuth.cache.capacity`, e.g., 10,000).
- **Time-to-Live (TTL):** Expiration duration for cached credentials (configured via `auth.dbAuth.cache.ttlSeconds`, e.g., 300 seconds).
- **Thread Safety:** Synchronized with `sync.RWMutex` to allow concurrent token checks across thousands of simultaneous connections.

```yaml
auth:
    dbEnabled: true
    dbAuth:
        cache:
            capacity: 10000
            ttlSeconds: 300
```

---

## 4. OpenAPI Specification Caching

The generated OpenAPI 3.0 JSON specification is cached in memory after its initial compilation.

- Requests to `/v1/openapi.json` and the interactive `/v1/docs` UI serve the pre-rendered specification with zero JSON serialization overhead.
- Schema mutations via DDL endpoints immediately invalidate the OpenAPI cache, triggering regeneration on the next access.
