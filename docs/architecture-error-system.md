# Error System Architecture

Conduit employs a centralized, RFC-style structured error system. All application errors, validation failures, database faults, and authorization rejections are formatted into a predictable, multi-format response envelope.

---

## Error Envelope Structure

Regardless of output format (JSON, XML, YAML, TOML, or CBOR), errors adhere to the standard envelope structure:

```json
{
    "error": {
        "status": 404,
        "title": "Not Found",
        "message": "The requested resource does not exist.",
        "details": ["Table 'books' contains no record with id '999'"],
        "documentation": "https://conduit.untapped.tech/docs/errors/not-found"
    }
}
```

### Core Envelope Fields

| Field           | Type                | Description                                                                       |
| --------------- | ------------------- | --------------------------------------------------------------------------------- |
| `status`        | `integer`           | HTTP status code matching the response header.                                    |
| `title`         | `string`            | Canonical human-readable error title.                                             |
| `message`       | `string`            | Localized or default descriptive explanation.                                     |
| `details`       | `array` of `string` | Specific diagnostic error messages, missing fields, or offending constraints.     |
| `documentation` | `string`            | Direct link to the official error documentation page explaining causes and fixes. |

---

## Multi-Format Error Serialization

Errors are serialized to match the client's requested format:

### YAML Error Example

```yaml
error:
    status: 400
    title: Bad Request
    message: Invalid column supplied in order parameter
    details:
        - Column 'invalid_col' does not exist in table 'books'
    documentation: https://conduit.untapped.tech/docs/errors/bad-request
```

### XML Error Example

```xml
<?xml version="1.0" encoding="UTF-8"?>
<response>
  <error>
    <status>403</status>
    <title>Forbidden</title>
    <message>You do not have permission to perform this action.</message>
    <documentation>https://conduit.untapped.tech/docs/errors/forbidden</documentation>
  </error>
</response>
```

---

## Error Catalog Index

Conduit classifies errors using standardized specifications:

| Status Code | Error Title            | Documentation Link                                             |
| ----------- | ---------------------- | -------------------------------------------------------------- |
| `400`       | Bad Request            | [400 Bad Request](errors/bad-request.md)                       |
| `401`       | Unauthorized           | [401 Unauthorized](errors/unauthorized.md)                     |
| `403`       | Forbidden              | [403 Forbidden](errors/forbidden.md)                           |
| `404`       | Not Found              | [404 Not Found](errors/not-found.md)                           |
| `405`       | Method Not Allowed     | [405 Method Not Allowed](errors/method-not-allowed.md)         |
| `409`       | Conflict               | [409 Conflict](errors/conflict.md)                             |
| `415`       | Unsupported Media Type | [415 Unsupported Media Type](errors/unsupported-media-type.md) |
| `422`       | Unprocessable Entity   | [422 Unprocessable Entity](errors/unprocessable-entity.md)     |
| `429`       | Too Many Requests      | [429 Too Many Requests](errors/too-many-requests.md)           |
| `500`       | Internal Server Error  | [500 Internal Server Error](errors/internal-server-error.md)   |
| `501`       | Not Implemented        | [501 Not Implemented](errors/not-implemented.md)               |
| `503`       | Service Unavailable    | [503 Service Unavailable](errors/service-unavailable.md)       |

For comprehensive troubleshooting details, visit the [Error Index](errors/index.md).
