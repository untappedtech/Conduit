package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/untappedtech/conduit/internal/db/impl"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

func boolPtr(b bool) *bool {
	return &b
}

type astTestCase struct {
	name          string
	ast           service.WhereExpr
	whereStr      string
	expectedCount int
	expectedNames []string
}

func getDialectTestCases() []astTestCase {
	return []astTestCase{
		{
			name: "Equality Comparison",
			ast: &service.ComparisonExpr{
				Column: "status",
				Op:     "=",
				Value:  "active",
			},
			whereStr:      "status = 'active'",
			expectedCount: 2,
			expectedNames: []string{"Alpha", "Gamma"},
		},
		{
			name: "Numeric Greater Than",
			ast: &service.ComparisonExpr{
				Column: "score",
				Op:     ">",
				Value:  50,
			},
			whereStr:      "score > 50",
			expectedCount: 2,
			expectedNames: []string{"Alpha", "Gamma"},
		},
		{
			name: "Logical Binary AND",
			ast: &service.BinaryExpr{
				Op: "AND",
				Left: &service.ComparisonExpr{
					Column: "status",
					Op:     "=",
					Value:  "active",
				},
				Right: &service.ComparisonExpr{
					Column: "score",
					Op:     ">=",
					Value:  150,
				},
			},
			whereStr:      "status = 'active' AND score >= 150",
			expectedCount: 1,
			expectedNames: []string{"Gamma"},
		},
		{
			name: "Logical Binary OR",
			ast: &service.BinaryExpr{
				Op: "OR",
				Left: &service.ComparisonExpr{
					Column: "status",
					Op:     "=",
					Value:  "inactive",
				},
				Right: &service.ComparisonExpr{
					Column: "status",
					Op:     "=",
					Value:  "pending",
				},
			},
			whereStr:      "(status = 'inactive' OR status = 'pending')",
			expectedCount: 2,
			expectedNames: []string{"Beta", "Delta"},
		},
		{
			name: "IN Set Expression",
			ast: &service.InExpr{
				Column: "status",
				Not:    false,
				Values: []any{"active", "pending"},
			},
			whereStr:      "status IN ('active', 'pending')",
			expectedCount: 3,
			expectedNames: []string{"Alpha", "Gamma", "Delta"},
		},
		{
			name: "NOT IN Set Expression",
			ast: &service.InExpr{
				Column: "status",
				Not:    true,
				Values: []any{"active", "inactive"},
			},
			whereStr:      "status NOT IN ('active', 'inactive')",
			expectedCount: 1,
			expectedNames: []string{"Delta"},
		},
		{
			name: "Logical NOT Negation",
			ast: &service.NotExpr{
				Expr: &service.ComparisonExpr{
					Column: "status",
					Op:     "=",
					Value:  "active",
				},
			},
			whereStr:      "NOT (status = 'active')",
			expectedCount: 2,
			expectedNames: []string{"Beta", "Delta"},
		},
		{
			name: "IS NOT NULL Check",
			ast: &service.ComparisonExpr{
				Column: "score",
				Op:     "IS NOT",
				Value:  nil,
			},
			whereStr:      "score IS NOT NULL",
			expectedCount: 4,
			expectedNames: []string{"Alpha", "Beta", "Gamma", "Delta"},
		},
	}
}

func setupTestItems(ctx context.Context, driver domain.DatabaseDriver) error {
	cols := []domain.ColumnDef{
		{Name: "id", Type: "INTEGER", PK: boolPtr(true), Autoincrement: boolPtr(true)},
		{Name: "name", Type: "VARCHAR(100)"},
		{Name: "status", Type: "VARCHAR(50)"},
		{Name: "score", Type: "INTEGER"},
	}

	if err := driver.CreateTable(ctx, "test_items", cols); err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	seedData := []map[string]any{
		{"name": "Alpha", "status": "active", "score": 100},
		{"name": "Beta", "status": "inactive", "score": 50},
		{"name": "Gamma", "status": "active", "score": 200},
		{"name": "Delta", "status": "pending", "score": 10},
	}

	for _, item := range seedData {
		if _, err := driver.Insert(ctx, "test_items", item); err != nil {
			return fmt.Errorf("seed insert failed for %v: %w", item, err)
		}
	}

	return nil
}

func TestPostgres_Integration(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	driver, err := impl.NewPostgresEngine(connStr)
	if err != nil {
		t.Fatalf("failed to initialize postgres engine: %v", err)
	}

	rawDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open raw postgres connection: %v", err)
	}
	defer rawDB.Close()

	if err := setupTestItems(ctx, driver); err != nil {
		t.Fatalf("setup test items failed: %v", err)
	}
	defer func() {
		_ = driver.DropTable(ctx, "test_items")
	}()

	dialect, ok := driver.(service.SQLDialect)
	if !ok {
		t.Fatalf("postgres driver does not implement service.SQLDialect")
	}

	testCases := getDialectTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Validate custom AST SQL generation with Postgres dialect ($1, $2, "ident")
			sqlWhere, args := tc.ast.ToSQL(dialect, 1)
			if sqlWhere == "" {
				t.Fatalf("generated empty WHERE clause for %s", tc.name)
			}

			// 2. Execute raw query against Postgres using the AST output
			query := fmt.Sprintf("SELECT name FROM %s WHERE %s ORDER BY id", dialect.QuoteIdent("test_items"), sqlWhere)
			rows, err := rawDB.QueryContext(ctx, query, args...)
			if err != nil {
				t.Fatalf("Postgres query failed for AST [%s] with args %v: %v", sqlWhere, args, err)
			}
			defer rows.Close()

			var names []string
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err != nil {
					t.Fatalf("scan error: %v", err)
				}
				names = append(names, name)
			}

			if len(names) != tc.expectedCount {
				t.Errorf("AST execution: expected %d rows, got %d (%v) for query [%s] args %v",
					tc.expectedCount, len(names), names, query, args)
			}

			// 3. Verify driver.List parses and executes the matching whereStr
			driverResults, err := driver.List(ctx, "test_items", domain.ListRequest{Where: tc.whereStr, Order: "id:asc"})
			if err != nil {
				t.Fatalf("driver.List failed for where %q: %v", tc.whereStr, err)
			}
			if len(driverResults) != tc.expectedCount {
				t.Errorf("driver.List: expected %d rows, got %d for where %q", tc.expectedCount, len(driverResults), tc.whereStr)
			}
		})
	}
}

func TestMySQL_Integration(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("failed to start mysql container: %v", err)
	}
	defer func() {
		if err := mysqlContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate mysql container: %v", err)
		}
	}()

	connStr, err := mysqlContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get mysql connection string: %v", err)
	}

	driver, err := impl.NewMySQLEngine(connStr)
	if err != nil {
		t.Fatalf("failed to initialize mysql engine: %v", err)
	}

	rawDB, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("failed to open raw mysql connection: %v", err)
	}
	defer rawDB.Close()

	if err := setupTestItems(ctx, driver); err != nil {
		t.Fatalf("setup test items failed: %v", err)
	}
	defer func() {
		_ = driver.DropTable(ctx, "test_items")
	}()

	dialect, ok := driver.(service.SQLDialect)
	if !ok {
		t.Fatalf("mysql driver does not implement service.SQLDialect")
	}

	testCases := getDialectTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Validate custom AST SQL generation with MySQL dialect (?, ?, `ident`)
			sqlWhere, args := tc.ast.ToSQL(dialect, 1)
			if sqlWhere == "" {
				t.Fatalf("generated empty WHERE clause for %s", tc.name)
			}

			// 2. Execute raw query against MySQL using the AST output
			query := fmt.Sprintf("SELECT name FROM %s WHERE %s ORDER BY id", dialect.QuoteIdent("test_items"), sqlWhere)
			rows, err := rawDB.QueryContext(ctx, query, args...)
			if err != nil {
				t.Fatalf("MySQL query failed for AST [%s] with args %v: %v", sqlWhere, args, err)
			}
			defer rows.Close()

			var names []string
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err != nil {
					t.Fatalf("scan error: %v", err)
				}
				names = append(names, name)
			}

			if len(names) != tc.expectedCount {
				t.Errorf("AST execution: expected %d rows, got %d (%v) for query [%s] args %v",
					tc.expectedCount, len(names), names, query, args)
			}

			// 3. Verify driver.List parses and executes the matching whereStr
			driverResults, err := driver.List(ctx, "test_items", domain.ListRequest{Where: tc.whereStr, Order: "id:asc"})
			if err != nil {
				t.Fatalf("driver.List failed for where %q: %v", tc.whereStr, err)
			}
			if len(driverResults) != tc.expectedCount {
				t.Errorf("driver.List: expected %d rows, got %d for where %q", tc.expectedCount, len(driverResults), tc.whereStr)
			}
		})
	}
}
