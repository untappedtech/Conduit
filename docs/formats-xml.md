# XML Format

Conduit features full XML decoding and custom schema-aware XML encoding. Requests and responses map relational data into clean, predictable XML element trees.

- **MIME Types:** `application/xml`, `text/xml`
- **Query Parameter:** `?format=xml`
- **Supported Operations:** Input (Decode) & Output (Encode)

---

## Sending XML Input

To submit XML data, set `Content-Type: application/xml` or `Content-Type: text/xml`:

### Single Record (POST / PUT)

```bash
curl -X POST http://localhost:8080/v1/books \
  -H "Content-Type: application/xml" \
  -d '<?xml version="1.0" encoding="UTF-8"?>
<book>
  <title>Foundation</title>
  <author>Isaac Asimov</author>
  <published_year>1951</published_year>
</book>'
```

Conduit's robust XML decoder parses records regardless of the enclosing root tag name (e.g. `<book>`, `<record>`, or `<row>`).

### Batch Ingestion (POST)

```bash
curl -X POST http://localhost:8080/v1/books \
  -H "Content-Type: application/xml" \
  -d '<?xml version="1.0" encoding="UTF-8"?>
<books>
  <book>
    <title>Foundation and Empire</title>
    <author>Isaac Asimov</author>
  </book>
  <book>
    <title>Second Foundation</title>
    <author>Isaac Asimov</author>
  </book>
</books>'
```

---

## Receiving XML Output

Request XML output using `?format=xml` or the `Accept: application/xml` header:

```bash
curl http://localhost:8080/v1/books?format=xml
```

### List Response

List responses are enveloped in `<response>` with child `<item>` elements matching the table rows:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<response>
  <item>
    <id>1</id>
    <title>Foundation</title>
    <author>Isaac Asimov</author>
    <published_year>1951</published_year>
  </item>
  <item>
    <id>2</id>
    <title>Foundation and Empire</title>
    <author>Isaac Asimov</author>
    <published_year>1952</published_year>
  </item>
</response>
```

### Single Record Response

Single record lookups (`GET /v1/books/1?format=xml`) render a concise document:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<response>
  <id>1</id>
  <title>Foundation</title>
  <author>Isaac Asimov</author>
  <published_year>1951</published_year>
</response>
```

---

## Key Characteristics

- **Schema Ordering:** Tag ordering strictly follows the database column definitions.
- **Null Values:** Nullable columns with `NULL` database values are omitted or represented with empty elements `<column></column>`.
