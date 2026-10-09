# NDJSON Format

NDJSON (Newline-Delimited JSON, also known as JSON Lines / JSONL) writes one valid JSON object per line separated by newline (`\n`) characters. It is the premier format for high-throughput streaming, log processing, and incremental ingestion pipelines.

- **MIME Type:** `application/x-ndjson`
- **Query Parameter:** `?format=ndjson`
- **Supported Operations:** Input (Decode) & Output (Encode)

---

## Sending NDJSON Input

To stream records into Conduit, set `Content-Type: application/x-ndjson`:

### Ingesting Multiple Records (POST)

```bash
curl -X POST http://localhost:8080/v1/metrics \
  -H "Content-Type: application/x-ndjson" \
  -d '{"sensor_id": "temp-01", "value": 21.4, "timestamp": "2026-10-08T12:00:00Z"}
{"sensor_id": "temp-02", "value": 22.8, "timestamp": "2026-10-08T12:00:00Z"}
{"sensor_id": "humidity-01", "value": 45.2, "timestamp": "2026-10-08T12:00:00Z"}'
```

Conduit iterates through the stream line-by-line, parsing and inserting each object.

---

## Receiving NDJSON Output

Request NDJSON responses with `?format=ndjson` or `Accept: application/x-ndjson`:

```bash
curl http://localhost:8080/v1/metrics?format=ndjson
```

### Response Example

```json
{"id":1,"sensor_id":"temp-01","value":21.4,"timestamp":"2026-10-08T12:00:00Z"}
{"id":2,"sensor_id":"temp-02","value":22.8,"timestamp":"2026-10-08T12:00:00Z"}
{"id":3,"sensor_id":"humidity-01","value":45.2,"timestamp":"2026-10-08T12:00:00Z"}
```

---

## Key Characteristics

- **Streaming Efficiency:** Clients can parse records incrementally line-by-line without buffering the entire HTTP payload into memory.
- **Full Query Support:** NDJSON output seamlessly respects `limit`, `offset`, `where`, and `order` query parameters.
