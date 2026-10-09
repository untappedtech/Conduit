# Database Configuration

The `database` configuration block defines how Conduit connects to your relational database backend.

You specify the driver identifier and the database-specific Data Source Name (DSN). Conduit passes the DSN directly to the chosen database engine driver.

---

## Configuration Settings

```yaml
database:
    driver: "sqlite"
    dsn: "./app.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
```

| Field    | Type     | Default    | Description                                                |
| -------- | -------- | ---------- | ---------------------------------------------------------- |
| `driver` | `string` | `"sqlite"` | Identifier of the database driver to use.                  |
| `dsn`    | `string` | `""`       | Data Source Name / connection string passed to the driver. |

---

## Supported Database Drivers

| Driver Identifier | Aliases             | Description                                                                 |
| ----------------- | ------------------- | --------------------------------------------------------------------------- |
| `sqlite`          | `sqlite3`           | Embedded file-based or in-memory SQLite (pure-Go via `modernc.org/sqlite`). |
| `postgres`        | `postgresql`, `pgx` | PostgreSQL relational database (via `pgx`).                                 |
| `mysql`           | `mariadb`           | MySQL and MariaDB servers (via `go-sql-driver/mysql`).                      |
| `sqlserver`       | `mssql`, `azuresql` | Microsoft SQL Server & Azure SQL (via `go-mssqldb`).                        |
| `libsql`          | `turso`             | libSQL edge database and Turso distributed cloud database.                  |
| `clickhouse`      | —                   | ClickHouse columnar analytics database (via native TCP protocol).           |
| `oracle`          | `godror`            | Oracle Database enterprise backend (pure-Go via `go-ora`).                  |
| `memory`          | `in-memory`         | Pure in-memory mock database for testing and ephemeral workloads.           |

---

## Driver Connection Examples

### 1. SQLite

```yaml
database:
    driver: "sqlite"
    dsn: "./app.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
```

### 2. PostgreSQL

```yaml
database:
    driver: "postgres"
    dsn: "postgres://username:secret@localhost:5432/my_database?sslmode=disable"
```

### 3. MySQL / MariaDB

```yaml
database:
    driver: "mysql"
    dsn: "username:secret@tcp(localhost:3306)/my_database?parseTime=true&charset=utf8mb4"
```

> **Recommendation:** Include `parseTime=true` so `DATETIME` and `TIMESTAMP` columns parse correctly as native Go timestamps.

### 4. Microsoft SQL Server

```yaml
database:
    driver: "sqlserver"
    dsn: "sqlserver://sa:StrongP@ssw0rd@localhost:1433?database=appdb&encrypt=disable"
```

### 5. libSQL / Turso

```yaml
database:
    driver: "libsql"
    dsn: "libsql://my-db-untapped.turso.io?authToken=eyJhbGciOi..."
```

### 6. ClickHouse

```yaml
database:
    driver: "clickhouse"
    dsn: "clickhouse://default:password@localhost:9000/default?dial_timeout=5s"
```

### 7. Oracle Database

```yaml
database:
    driver: "oracle"
    dsn: "oracle://system:oracle@localhost:1521/XE"
```

### 8. In-Memory Mode

```yaml
database:
    driver: "memory"
    dsn: ""
```

The memory driver creates a temporary table store in RAM without persisting data to disk. It is ideal for local unit tests and CI/CD pipelines.
