# Policy Configuration

The `policy` configuration block controls global unauthenticated access permissions. These settings govern what operations anonymous clients (requests without authorization tokens) are permitted to execute.

---

## Configuration Settings

```yaml
policy:
    publicReads: true
    publicWrites: false
    publicMutation: false
```

| Field            | Type      | Default | Description                                                                                                                        |
| ---------------- | --------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `publicReads`    | `boolean` | `true`  | Allows unauthenticated `GET` and `HEAD` operations on table endpoints and schema introspection.                                    |
| `publicWrites`   | `boolean` | `false` | Allows unauthenticated data mutation (`POST` insert, `PUT` replace, `PATCH` update, and `DELETE` record).                          |
| `publicMutation` | `boolean` | `false` | Allows unauthenticated schema-level mutations (`POST /v1/schema/<table>` create table and `DELETE /v1/schema/<table>` drop table). |

---

## Common Security Profiles

### 1. Completely Open (Development & Prototyping)

Allows any client to read, write data, and create/drop tables without an API key or token:

```yaml
policy:
    publicReads: true
    publicWrites: true
    publicMutation: true
```

### 2. Public Read-Only (Content & Catalogs)

Allows public visitors to query tables, while write and schema operations require an authenticated token:

```yaml
policy:
    publicReads: true
    publicWrites: false
    publicMutation: false
```

### 3. Fully Locked Down (Internal & Enterprise APIs)

Rejects all requests unless a valid authorization token is presented:

```yaml
policy:
    publicReads: false
    publicWrites: false
    publicMutation: false
```

---

## Enforcement Behavior

When a policy flag is set to `false`, incoming requests without valid authentication tokens receive a `401 Unauthorized` error response.

If a client supplies a valid token, the request is evaluated against the client's assigned role permissions (`readOnly`, `readWrite`, or `admin`), overriding the public policy restrictions.
