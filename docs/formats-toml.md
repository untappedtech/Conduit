# TOML Format

TOML (Tom's Obvious Minimal Language) is designed for clear, unambiguous representation of key/value pairs and tables. Conduit supports TOML for both request decoding and response encoding.

- **MIME Types:** `application/toml`, `text/toml`
- **Query Parameter:** `?format=toml`
- **Supported Operations:** Input (Decode) & Output (Encode)

---

## Sending TOML Input

Set `Content-Type: application/toml`:

### Create Single Record (POST)

```bash
curl -X POST http://localhost:8080/v1/projects \
  -H "Content-Type: application/toml" \
  -d '
name = "Apollo"
budget = 250000.50
lead = "Gene Kranz"
active = true
'
```

### Batch Insertion (POST)

Using TOML array of tables:

```bash
curl -X POST http://localhost:8080/v1/projects \
  -H "Content-Type: application/toml" \
  -d '
[[projects]]
name = "Gemini"
budget = 120000.00
lead = "Gus Grissom"

[[projects]]
name = "Mercury"
budget = 80000.00
lead = "Alan Shepard"
'
```

---

## Receiving TOML Output

Request TOML output with `?format=toml` or `Accept: application/toml`:

```bash
curl http://localhost:8080/v1/projects?format=toml
```

### List Response

List responses are encoded using TOML array of tables `[[projects]]` matching the table name:

```toml
[[projects]]
id = 1
name = "Apollo"
budget = 250000.5
lead = "Gene Kranz"
active = true

[[projects]]
id = 2
name = "Gemini"
budget = 120000.0
lead = "Gus Grissom"
active = false
```

### Single Record Response

```toml
id = 1
name = "Apollo"
budget = 250000.5
lead = "Gene Kranz"
active = true
```

---

## Key Characteristics

- **Semantic Datatypes:** Integers, floats, booleans, and ISO-8601 strings serialize into native TOML representations.
- **Top-Level Grouping:** Multiple records automatically generate `[[<tableName>]]` arrays of tables.
