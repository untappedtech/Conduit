# YAML Format

YAML (YAML Ain't Markup Language) provides human-readable, indentation-based data serialization. Conduit natively parses and renders YAML for all endpoints.

- **MIME Types:** `application/x-yaml`, `text/yaml`
- **Query Parameter:** `?format=yaml` or `?format=yml`
- **Supported Operations:** Input (Decode) & Output (Encode)

---

## Sending YAML Input

To send YAML payloads to Conduit, set `Content-Type: application/x-yaml`:

### Create Single Record (POST)

```bash
curl -X POST http://localhost:8080/v1/servers \
  -H "Content-Type: application/x-yaml" \
  -d '
hostname: prod-web-01
ip_address: 10.0.1.25
cores: 8
active: true
'
```

### Batch Insertion (POST)

```bash
curl -X POST http://localhost:8080/v1/servers \
  -H "Content-Type: application/x-yaml" \
  -d '
- hostname: prod-db-01
  ip_address: 10.0.1.30
  cores: 16
  active: true
- hostname: prod-cache-01
  ip_address: 10.0.1.40
  cores: 4
  active: true
'
```

---

## Receiving YAML Output

Request YAML formatting using `?format=yaml` or the `Accept: application/x-yaml` header:

```bash
curl http://localhost:8080/v1/servers?format=yaml
```

### List Response

```yaml
- id: 1
  hostname: prod-web-01
  ip_address: 10.0.1.25
  cores: 8
  active: true
- id: 2
  hostname: prod-db-01
  ip_address: 10.0.1.30
  cores: 16
  active: true
```

### Single Record Response

```yaml
id: 1
hostname: prod-web-01
ip_address: 10.0.1.25
cores: 8
active: true
```

---

## Key Characteristics

- **Schema Ordering:** YAML keys preserve the underlying column order of the SQL table.
- **Data Types:** Numbers, booleans, strings, and null scalars map directly to native YAML scalar types.
