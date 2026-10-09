# 415 Unsupported Media Type

The request's `Content-Type` format is not supported by Conduit's decoder engine.

---

## What This Error Means

Conduit received a payload with a `Content-Type` header that it does not recognize or support for input decoding.

---

## Supported Input Formats

Conduit supports incoming request bodies in the following formats:

- **JSON:** `application/json`
- **XML:** `application/xml`, `text/xml`
- **YAML:** `application/x-yaml`, `text/yaml`
- **TOML:** `application/toml`, `text/toml`
- **NDJSON:** `application/x-ndjson`
- **CSV:** `text/csv`
- **CBOR:** `application/cbor`

---

## Common Causes

- Setting `Content-Type: multipart/form-data` or `application/x-www-form-urlencoded`.
- Missing or mistyped `Content-Type` header (e.g. `text/plain` when sending JSON).
- Sending binary formats other than CBOR.

---

## How to Fix It

1. **Set Correct Content-Type:** Set the `Content-Type` header to one of the 7 supported MIME types.
2. **Format Payload to Match:** Ensure the request body matches the specified `Content-Type`.

---

## Example Request & Response

### Request

```http
POST /v1/books
Content-Type: application/zip

[binary data]
```

### Response

```json
{
    "error": {
        "status": 415,
        "title": "Unsupported Media Type",
        "message": "The payload content type 'application/zip' is not supported.",
        "details": ["Supported formats: json, xml, yaml, toml, ndjson, csv, cbor"],
        "documentation": "https://conduit.untapped.tech/docs/errors/unsupported-media-type"
    }
}
```

---

[← Back to Error Index](index.md)
