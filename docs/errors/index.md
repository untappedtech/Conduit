# Conduit Error Documentation

This catalog provides documentation, root causes, and remediation guidance for all HTTP error responses emitted by Conduit.

Every Conduit error response includes a structured RFC-style envelope with the HTTP status, error title, message, specific details, and a link to this catalog.

---

## Error Catalog

| Status Code | Error Title            | Description                                                                   | Link                                                    |
| :---------: | ---------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------- |
|   **400**   | Bad Request            | The request was malformed or referenced invalid schema columns.               | [400 Bad Request](bad-request.md)                       |
|   **401**   | Unauthorized           | Authentication is required to access the requested resource.                  | [401 Unauthorized](unauthorized.md)                     |
|   **403**   | Forbidden              | The authenticated client does not have sufficient role privileges.            | [403 Forbidden](forbidden.md)                           |
|   **404**   | Not Found              | The requested table or record primary key does not exist.                     | [404 Not Found](not-found.md)                           |
|   **405**   | Method Not Allowed     | The HTTP method is not permitted on this endpoint.                            | [405 Method Not Allowed](method-not-allowed.md)         |
|   **409**   | Conflict               | The operation violates a unique constraint or primary key.                    | [409 Conflict](conflict.md)                             |
|   **415**   | Unsupported Media Type | The `Content-Type` header specifies an unsupported payload format.            | [415 Unsupported Media Type](unsupported-media-type.md) |
|   **422**   | Unprocessable Entity   | Payload validation failed against table column rules or invalid query syntax. | [422 Unprocessable Entity](unprocessable-entity.md)     |
|   **429**   | Too Many Requests      | Rate limit exceeded.                                                          | [429 Too Many Requests](too-many-requests.md)           |
|   **500**   | Internal Server Error  | An unexpected server or database driver failure occurred.                     | [500 Internal Server Error](internal-server-error.md)   |
|   **501**   | Not Implemented        | The requested database driver feature or function is not implemented.         | [501 Not Implemented](not-implemented.md)               |
|   **503**   | Service Unavailable    | The database connection is offline or temporarily unavailable.                | [503 Service Unavailable](service-unavailable.md)       |

---

[← Back to Documentation Overview](../index.md)
