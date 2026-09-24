package service_test

import (
	"testing"

	"github.com/untappedtech/conduit/internal/db"
	"github.com/untappedtech/conduit/internal/db/impl"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

func TestDriverDialects_QuoteAndPlaceholder(t *testing.T) {
	sqlite := &impl.SQLiteEngine{}
	postgres := &impl.PostgresEngine{}
	mysql := &impl.MySQLEngine{}
	sqlserver := &impl.SQLServerEngine{}
	libsql := &impl.LibSQLEngine{}
	clickhouse := &impl.ClickHouseEngine{}
	oracle := &impl.OracleEngine{}

	tests := []struct {
		name            string
		dialect         service.SQLDialect
		colName         string
		expectedQuote   string
		paramIndex      int
		expectedPlace   string
	}{
		{"SQLite", sqlite, "user_name", `"user_name"`, 1, "?"},
		{"SQLite Escaped", sqlite, `user"name`, `"user""name"`, 2, "?"},
		{"PostgreSQL", postgres, "user_name", `"user_name"`, 3, "$3"},
		{"MySQL", mysql, "user_name", "`user_name`", 1, "?"},
		{"MySQL Escaped", mysql, "user`name", "`user``name`", 1, "?"},
		{"SQL Server", sqlserver, "user_name", "[user_name]", 4, "@p4"},
		{"SQL Server Escaped", sqlserver, "user]name", "[user]]name]", 4, "@p4"},
		{"libSQL", libsql, "user_name", `"user_name"`, 1, "?"},
		{"ClickHouse", clickhouse, "user_name", "`user_name`", 1, "?"},
		{"Oracle", oracle, "user_name", `"user_name"`, 5, ":5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			quoted := tc.dialect.QuoteIdent(tc.colName)
			if quoted != tc.expectedQuote {
				t.Errorf("QuoteIdent(%q) = %q, want %q", tc.colName, quoted, tc.expectedQuote)
			}

			placeholder := tc.dialect.Placeholder(tc.paramIndex)
			if placeholder != tc.expectedPlace {
				t.Errorf("Placeholder(%d) = %q, want %q", tc.paramIndex, placeholder, tc.expectedPlace)
			}
		})
	}
}

func TestDriverDialects_WhereCompilation(t *testing.T) {
	cols := []domain.ColumnDef{
		{Name: "id", Type: "INTEGER"},
		{Name: "status", Type: "TEXT"},
		{Name: "score", Type: "INTEGER"},
	}

	clickhouse := &impl.ClickHouseEngine{}
	oracle := &impl.OracleEngine{}
	libsql := &impl.LibSQLEngine{}

	whereClause := "status = 'active' AND score >= 100"

	// Test ClickHouse SQL generation
	chSQL, chArgs, err := service.ParseWhereSQL(whereClause, cols, clickhouse, 1)
	if err != nil {
		t.Fatalf("ClickHouse ParseWhereSQL error: %v", err)
	}
	expectedChSQL := "(`status` = ? AND `score` >= ?)"
	if chSQL != expectedChSQL {
		t.Errorf("ClickHouse SQL = %q, want %q", chSQL, expectedChSQL)
	}
	if len(chArgs) != 2 || chArgs[0] != "active" || chArgs[1] != int64(100) {
		t.Errorf("ClickHouse args = %v, want ['active', 100]", chArgs)
	}

	// Test Oracle SQL generation
	oraSQL, oraArgs, err := service.ParseWhereSQL(whereClause, cols, oracle, 1)
	if err != nil {
		t.Fatalf("Oracle ParseWhereSQL error: %v", err)
	}
	expectedOraSQL := `("status" = :1 AND "score" >= :2)`
	if oraSQL != expectedOraSQL {
		t.Errorf("Oracle SQL = %q, want %q", oraSQL, expectedOraSQL)
	}
	if len(oraArgs) != 2 || oraArgs[0] != "active" || oraArgs[1] != int64(100) {
		t.Errorf("Oracle args = %v, want ['active', 100]", oraArgs)
	}

	// Test libSQL SQL generation
	libSQL, libArgs, err := service.ParseWhereSQL(whereClause, cols, libsql, 1)
	if err != nil {
		t.Fatalf("libSQL ParseWhereSQL error: %v", err)
	}
	expectedLibSQL := `("status" = ? AND "score" >= ?)`
	if libSQL != expectedLibSQL {
		t.Errorf("libSQL SQL = %q, want %q", libSQL, expectedLibSQL)
	}
	if len(libArgs) != 2 {
		t.Errorf("libSQL args len = %d, want 2", len(libArgs))
	}
}

func TestDriverFactory_UnsupportedDriver(t *testing.T) {
	cfg := &domain.ServerConfig{}
	cfg.Database.Driver = "unsupported_xyz"
	cfg.Database.DSN = "some_dsn"

	_, err := db.NewDatabase(cfg)
	if err == nil {
		t.Errorf("expected error for unsupported driver, got nil")
	}
}

func TestDriverFactory_Aliases(t *testing.T) {
	memoryAliases := []string{"memory", "mem", "in-memory", "inmemory", "MEMORY", " In-Memory "}

	for _, alias := range memoryAliases {
		cfg := &domain.ServerConfig{}
		cfg.Database.Driver = alias

		driver, err := db.NewDatabase(cfg)
		if err != nil {
			t.Errorf("expected alias %q to succeed, got error: %v", alias, err)
		}
		if driver == nil {
			t.Errorf("expected non-nil driver for alias %q", alias)
		}
	}
}
