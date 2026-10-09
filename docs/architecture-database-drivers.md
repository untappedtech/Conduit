# Database Drivers Architecture

Conduit achieves complete database portability by decoupling high-level REST handling from SQL execution through a unified database driver interface.

Every supported database implements this interface, translating standard CRUD and schema operations into dialect-accurate SQL.

---

## Driver Interface

All drivers implement the core `Database` interface (`internal/domain/database.go`):

```go
type Database interface {
    // Schema Operations
    ListTables(ctx context.Context) ([]string, error)
    GetSchema(ctx context.Context, table string) ([]ColumnDef, error)
    CreateTable(ctx context.Context, table string, columns []ColumnDef) error
    DropTable(ctx context.Context, table string) error

    // CRUD Operations
    Insert(ctx context.Context, table string, record map[string]any) (map[string]any, error)
    Get(ctx context.Context, table string, id any) (map[string]any, error)
    Update(ctx context.Context, table string, id any, record map[string]any) (map[string]any, error)
    Patch(ctx context.Context, table string, id any, record map[string]any) (map[string]any, error)
    Delete(ctx context.Context, table string, id any) error
    List(ctx context.Context, table string, req ListRequest) ([]map[string]any, error)

    // Lifecycle
    Close() error
}
```

---

## Supported Database Backends

| Engine              | Identifier / Aliases             | Driver Implementation Package       | Quoting Style  | Parameter Placeholders |
| ------------------- | -------------------------------- | ----------------------------------- | -------------- | ---------------------- |
| **SQLite**          | `sqlite`, `sqlite3`              | `modernc.org/sqlite` (Pure-Go)      | `"column"`     | `?`                    |
| **PostgreSQL**      | `postgres`, `postgresql`, `pgx`  | `jackc/pgx/v5`                      | `"column"`     | `$1, $2, $3`           |
| **MySQL / MariaDB** | `mysql`, `mariadb`               | `go-sql-driver/mysql`               | `` `column` `` | `?`                    |
| **SQL Server**      | `sqlserver`, `mssql`, `azuresql` | `microsoft/go-mssqldb`              | `[column]`     | `@p1, @p2, @p3`        |
| **libSQL / Turso**  | `libsql`, `turso`                | `tursodatabase/go-libsql`           | `"column"`     | `?`                    |
| **ClickHouse**      | `clickhouse`                     | `ClickHouse/clickhouse-go/v2`       | `` `column` `` | `?`                    |
| **Oracle Database** | `oracle`, `godror`               | `sijms/go-ora/v2` (Pure-Go)         | `"COLUMN"`     | `:p1, :p2, :p3`        |
| **Memory**          | `memory`, `in-memory`            | Pure Go thread-safe in-memory store | None           | In-memory evaluation   |

---

## Dialect Handling

Different relational database engines require distinct syntax for identifier quoting, parameter placeholders, pagination, and data type mapping:

### 1. Identifier Quoting

- **SQLite / Postgres / libSQL:** Standard double quotes (`"users"`, `"created_at"`)
- **MySQL / ClickHouse:** Backticks (`` `users` ``, `` `created_at` ``)
- **SQL Server:** Square brackets (`[users]`, `[created_at]`)
- **Oracle:** Double quotes uppercase (`"USERS"`, `"CREATED_AT"`)

### 2. Parameter Placeholders

- **Positional `?`:** Used for SQLite, MySQL, libSQL, and ClickHouse.
- **Numbered `$N`:** Used for PostgreSQL (`$1`, `$2`).
- **Named `@pN`:** Used for SQL Server (`@p1`, `@p2`).
- **Colon-prefixed `:pN`:** Used for Oracle (`:p1`, `:p2`).

### 3. Pagination Dialects

- **`LIMIT ... OFFSET ...`:** Standard syntax for SQLite, Postgres, MySQL, libSQL, and ClickHouse.
- **`OFFSET ... ROWS FETCH NEXT ... ROWS ONLY`:** SQL standard syntax required by SQL Server and Oracle.

---

## Adding a Custom Driver

To add support for a new database engine:

1. Create a new file in `internal/db/impl/<engine>.go`.
2. Implement the `domain.Database` interface methods.
3. Register the new driver in `internal/db/factory.go`.
4. Run dialect and CRUD test suites against the new engine.
