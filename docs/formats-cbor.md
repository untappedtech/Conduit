# CBOR Format

CBOR (Concise Binary Object Representation, RFC 8949) is a binary data serialization format designed for high performance, compact transmission size, and minimal parsing overhead. It is ideal for microservices, mobile apps, and resource-constrained environments.

- **MIME Type:** `application/cbor`
- **Query Parameter:** `?format=cbor`
- **Supported Operations:** Input (Decode) & Output (Encode)

---

## Sending CBOR Input

Send binary CBOR records by setting `Content-Type: application/cbor`:

```bash
# Using a binary file or raw payload
curl -X POST http://localhost:8080/v1/telemetry \
  -H "Content-Type: application/cbor" \
  --data-binary @telemetry_payload.cbor
```

Conduit decodes the CBOR maps and arrays directly into schema-aligned data types.

---

## Receiving CBOR Output

Request CBOR responses by passing `?format=cbor` or setting `Accept: application/cbor`:

```bash
curl http://localhost:8080/v1/telemetry?format=cbor --output response.cbor
```

Or inspect using tools like `cbor-diag` / `cbor2diag`:

```bash
curl -s http://localhost:8080/v1/telemetry?format=cbor | cbor2diag.py
```

### Decoded Logical Representation

```json
[
    {
        "id": 1,
        "device_uuid": "f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
        "voltage": 3.29,
        "active": true
    }
]
```

---

## Key Characteristics

- **Deterministic Canonical Encoding:** Output bytes are generated canonically for reliable cryptographic hashing and caching.
- **Ultra-Low Overhead:** Eliminates text parsing and string quoting overhead, saving bandwidth and CPU cycles.
