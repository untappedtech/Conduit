# Configuration Overview

Conduit is configured via a single configuration file. It supports four file formats: **JSON**, **YAML**, **TOML**, and **XML**.

The configuration file dictates HTTP listener binding, base routing paths, operational database drivers and credentials, unauthenticated security policies, authentication mechanisms, and OpenAPI metadata.

---

## File Detection Order

When Conduit starts without an explicit `--config` flag, it searches the current working directory in the following order:

1. `config.json`
2. `config.yaml`
3. `config.yml`
4. `config.toml`
5. `config.xml`

You can also specify a custom configuration file path using the `-c` or `--config` flag:

```bash
conduit --config /etc/conduit/production.yaml
```

---

## Generating Default Configurations

To generate a fully commented starter configuration in your format of choice, use the `--generate-config` (`-g`) flag:

```bash
conduit --generate-config yaml   # Creates config.yaml
conduit --generate-config json   # Creates config.json
conduit --generate-config toml   # Creates config.toml
conduit --generate-config xml    # Creates config.xml
```

---

## Configuration Sections

The configuration file is divided into five core sections:

```yaml
server:
    # Network binding, ports, routing base path, and query limits

database:
    # Operational database engine driver and DSN connection string

policy:
    # Unauthenticated public read, write, and schema mutation flags

auth:
    # Token-based authentication (Environment, Database, or No-Op)

openapi:
    # Interactive documentation and OpenAPI 3.0 specification metadata
```

- [Server Settings](config-server.md) — Host, port, base path, and default pagination limit.
- [Database Settings](config-database.md) — Drivers (SQLite, PostgreSQL, MySQL, SQL Server, libSQL, ClickHouse, Oracle, Memory) and DSN syntax.
- [Policy Settings](config-policy.md) — Public access permissions (`publicReads`, `publicWrites`, `publicMutation`).
- [Authorization Overview](config-auth.md) — Auth strategies, token extractors, and role enforcement.
    - [Environment Auth](config-auth-env.md)
    - [Database Auth](config-auth-db.md)
    - [No-Op Auth](config-auth-noop.md)

---

## Complete Configuration Example (YAML)

```yaml
server:
    host: "0.0.0.0"
    port: 8080
    base_path: "/v1/"
    default_limit: 50

database:
    driver: "sqlite"
    dsn: "./app.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"

policy:
    publicReads: true
    publicWrites: true
    publicMutation: true

auth:
    environmentEnabled: false
    dbEnabled: false
    dbAuth:
        driver: "sqlite"
        dsn: "./auth.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
        table: "users"
        tokenColumn: "token"
        roleColumn: "role"
        roles:
            admin: "admin"
            readWrite: "rw"
            readOnly: "ro"
        cache:
            capacity: 10000
            ttlSeconds: 300

openapi:
    title: "Conduit Operational API"
    version: "1.0.0"
    description: "REST Engine automatically generated from relational SQL schema."
    contact:
        name: "Dev Team"
        email: "dev@untapped.tech"
        url: "https://untapped.tech"
    license:
        name: "MIT"
        url: "https://opensource.org/licenses/MIT"
    servers:
        - url: "http://localhost:8080/v1"
          description: "Local development server"
```

---

## Complete Configuration Example (JSON)

```json
{
    "server": {
        "host": "0.0.0.0",
        "port": 8080,
        "base_path": "/v1/",
        "default_limit": 50
    },
    "database": {
        "driver": "postgres",
        "dsn": "postgres://user:password@localhost:5432/mydb?sslmode=disable"
    },
    "policy": {
        "publicReads": true,
        "publicWrites": false,
        "publicMutation": false
    },
    "auth": {
        "environmentEnabled": true,
        "dbEnabled": false
    },
    "openapi": {
        "title": "Conduit Production API",
        "version": "1.0.0",
        "description": "PostgreSQL API exposed through Conduit"
    }
}
```
