# CSV Format

CSV (Comma-Separated Values) is the standard tabular data exchange format for business analytics, spreadsheets (Excel, Google Sheets), and database import utilities. Conduit provides first-class support for both importing and exporting CSV data.

- **MIME Type:** `text/csv`
- **Query Parameter:** `?format=csv`
- **Supported Operations:** Input (Decode) & Output (Encode)

---

## Sending CSV Input

To import CSV records directly into a table, set `Content-Type: text/csv`. The first line must contain column header names matching the target table:

### Bulk Insert (POST)

```bash
curl -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: text/csv" \
  -d 'name,department,salary
Alice Johnson,Engineering,115000
Bob Smith,Marketing,95000
Carol Danvers,Operations,105000'
```

Conduit parses header rows, casts field values to the underlying database column types, and inserts records in order.

---

## Receiving CSV Output

Export table rows directly as CSV using `?format=csv` or `Accept: text/csv`:

```bash
curl http://localhost:8080/v1/employees?format=csv
```

### Export Filtered & Sorted Data

```bash
curl "http://localhost:8080/v1/employees?where=salary%20%3E%3D%20100000&order=salary:desc&format=csv"
```

### Response Example

```csv
id,name,department,salary
1,Alice Johnson,Engineering,115000
3,Carol Danvers,Operations,105000
```

---

## Key Characteristics

- **Schema-Ordered Headers:** CSV column headers and row cells follow the column order established in the database schema.
- **Escape Handling:** Values containing commas, line breaks, or quotation marks are properly escaped adhering to RFC 4180 standards.
