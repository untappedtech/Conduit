# Multi-Format Serialization Engine

Conduit's multi-format serialization engine allows any relational SQL table to be queried or mutated across 7 formats: **JSON, XML, YAML, TOML, NDJSON, CSV, and CBOR**.

---

## Architecture Design

The engine is split into two complementary halves:

1. **Decoders:** Turn raw HTTP request payloads into standardized Go data structures (`map[string]any` or `[]map[string]any`).
2. **Encoders:** Transform Go data structures into serialized formats adhering to table schema column ordering.

```
Incoming Request Body           Go In-Memory Record              Outgoing Response Body
┌───────────────────────┐       ┌───────────────────────┐       ┌───────────────────────┐
│ JSON / XML / YAML /   │ ===>  │  map[string]any       │ ===>  │ JSON / XML / YAML /   │
│ TOML / NDJSON / CSV / │       │                       │       │ TOML / NDJSON / CSV / │
│ CBOR                  │       │  []map[string]any     │       │ CBOR                  │
└───────────────────────┘       └───────────────────────┘       └───────────────────────┘
     Decoder Layer                    Service Core                     Encoder Layer
```

---

## Decoder Details

Decoders inspect the `Content-Type` header:

```go
switch mediaType {
case "application/json":
    // JSON Object or Array Decoder
case "application/xml", "text/xml":
    // Schema-aware XML Document Decoder
case "application/x-yaml", "text/yaml":
    // YAML Stream Decoder
case "application/toml", "text/toml":
    // TOML Document / Table Decoder
case "application/x-ndjson":
    // Stream-based Line-by-Line JSON Decoder
case "text/csv":
    // Header-mapped CSV Tabular Decoder
case "application/cbor":
    // Binary Canonical CBOR Decoder
}
```

### Flexible Payload Normalization

Regardless of whether a single object or an array of items is submitted, Conduit normalizes the input into records and executes validation against the database column definitions.

---

## Encoder Details

The encoder constructs output representations that preserve database schema column order:

### 1. Schema-Ordered Key Sorting

Standard map serialization in Go produces non-deterministic key ordering. Conduit's encoders consult the active table schema and sort output fields so primary keys and main columns appear first, matching the exact order defined in SQL DDL.

### 2. Format Negotiation Hierarchy

1. **`?format=` query parameter:** Explicit user command.
2. **`Accept` header:** Standard HTTP negotiation.
3. **Echo input format:** Defaults response to the same format as request body.
4. **Fallback:** `application/json`.

---

## Performance Considerations

- **Streaming Formats:** NDJSON streams line-by-line directly to the HTTP response writer without allocating large contiguous buffers.
- **Binary Speed:** CBOR bypasses string parsing and quote escaping entirely, providing the lowest latency and smallest bandwidth footprint.
- **Cached Encoders:** Standard encoder instances are instantiated once and reused safely across concurrent goroutines.
