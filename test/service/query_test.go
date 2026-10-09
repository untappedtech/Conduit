package service_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

func sampleColumns() []domain.ColumnDef {
	return []domain.ColumnDef{
		{Name: "id", Type: "INTEGER"},
		{Name: "name", Type: "TEXT"},
		{Name: "age", Type: "INTEGER"},
		{Name: "status", Type: "TEXT"},
		{Name: "active", Type: "BOOLEAN"},
	}
}

func TestValidateOrder(t *testing.T) {
	cols := sampleColumns()

	tests := []struct {
		input    string
		wantCol  string
		wantDesc bool
		wantErr  bool
	}{
		{"", "", false, false},
		{"name", "name", false, false},
		{"NAME", "name", false, false},
		{"name:asc", "name", false, false},
		{"name:desc", "name", true, false},
		{"age:DESC", "age", true, false},
		{"invalid_col", "", false, true},
		{"name:invalid", "", false, true},
		{"name:asc:extra", "", false, true},
	}

	for _, tc := range tests {
		col, desc, err := service.ValidateOrder(tc.input, cols)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ValidateOrder(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ValidateOrder(%q) unexpected error: %v", tc.input, err)
			}
			if col != tc.wantCol || desc != tc.wantDesc {
				t.Errorf("ValidateOrder(%q) = (%q, %v), want (%q, %v)", tc.input, col, desc, tc.wantCol, tc.wantDesc)
			}
		}
	}
}

func TestParseWhere_ToSQL(t *testing.T) {
	cols := sampleColumns()

	sqliteDialect := service.SimpleDialect{
		QuoteIdentFunc:  func(col string) string { return `"` + strings.ReplaceAll(col, `"`, `""`) + `"` },
		PlaceholderFunc: func(i int) string { return "?" },
	}
	postgresDialect := service.SimpleDialect{
		QuoteIdentFunc:  func(col string) string { return `"` + strings.ReplaceAll(col, `"`, `""`) + `"` },
		PlaceholderFunc: func(i int) string { return fmt.Sprintf("$%d", i) },
	}
	sqlserverDialect := service.SimpleDialect{
		QuoteIdentFunc:  func(col string) string { return `[` + strings.ReplaceAll(col, `]`, `]]`) + `]` },
		PlaceholderFunc: func(i int) string { return fmt.Sprintf("@p%d", i) },
	}
	mysqlDialect := service.SimpleDialect{
		QuoteIdentFunc:  func(col string) string { return "`" + strings.ReplaceAll(col, "`", "``") + "`" },
		PlaceholderFunc: func(i int) string { return "?" },
	}

	tests := []struct {
		where        string
		dialect      service.SQLDialect
		startIndex   int
		expectedSQL  string
		expectedArgs []any
	}{
		{
			where:        "age > 21",
			dialect:      sqliteDialect,
			startIndex:   1,
			expectedSQL:  `"age" > ?`,
			expectedArgs: []any{int64(21)},
		},
		{
			where:        "status = 'active' AND age >= 18",
			dialect:      postgresDialect,
			startIndex:   1,
			expectedSQL:  `("status" = $1 AND "age" >= $2)`,
			expectedArgs: []any{"active", int64(18)},
		},
		{
			where:        "status = 'active' OR (age < 30 AND active = true)",
			dialect:      sqlserverDialect,
			startIndex:   1,
			expectedSQL:  `([status] = @p1 OR ([age] < @p2 AND [active] = @p3))`,
			expectedArgs: []any{"active", int64(30), true},
		},
		{
			where:        "name LIKE '%test%'",
			dialect:      mysqlDialect,
			startIndex:   1,
			expectedSQL:  "`name` LIKE ?",
			expectedArgs: []any{"%test%"},
		},
		{
			where:        "status IN ('active', 'pending')",
			dialect:      sqliteDialect,
			startIndex:   1,
			expectedSQL:  `"status" IN (?, ?)`,
			expectedArgs: []any{"active", "pending"},
		},
		{
			where:        "status IS NULL",
			dialect:      postgresDialect,
			startIndex:   1,
			expectedSQL:  `"status" IS NULL`,
			expectedArgs: nil,
		},
	}

	for _, tc := range tests {
		sql, args, err := service.ParseWhereSQL(tc.where, cols, tc.dialect, tc.startIndex)
		if err != nil {
			t.Fatalf("ParseWhereSQL(%q) error: %v", tc.where, err)
		}
		if sql != tc.expectedSQL {
			t.Errorf("ParseWhereSQL(%q) SQL = %q, want %q", tc.where, sql, tc.expectedSQL)
		}
		if len(args) != len(tc.expectedArgs) {
			t.Fatalf("ParseWhereSQL(%q) args len = %d, want %d", tc.where, len(args), len(tc.expectedArgs))
		}
		for i := range args {
			if args[i] != tc.expectedArgs[i] {
				t.Errorf("ParseWhereSQL(%q) args[%d] = %v, want %v", tc.where, i, args[i], tc.expectedArgs[i])
			}
		}
	}
}

func TestParseWhere_EvalMemory(t *testing.T) {
	cols := sampleColumns()

	rowMatch := map[string]any{
		"id":     1,
		"name":   "Alice Wonderland",
		"age":    25,
		"status": "active",
		"active": true,
	}

	rowNonMatch := map[string]any{
		"id":     2,
		"name":   "Bob Builder",
		"age":    40,
		"status": "inactive",
		"active": false,
	}

	tests := []struct {
		where     string
		matchRow1 bool
		matchRow2 bool
	}{
		{"age > 30", false, true},
		{"age <= 25", true, false},
		{"status = 'active'", true, false},
		{"name LIKE '%wonder%'", true, false},
		{"status IN ('active', 'pending')", true, false},
		{"status != 'active'", false, true},
		{"active = true", true, false},
		{"status = 'active' AND age < 30", true, false},
		{"status = 'active' OR age > 35", true, true},
	}

	for _, tc := range tests {
		expr, err := service.ParseWhere(tc.where, cols)
		if err != nil {
			t.Fatalf("ParseWhere(%q) error: %v", tc.where, err)
		}
		if res1 := expr.Eval(rowMatch); res1 != tc.matchRow1 {
			t.Errorf("Eval(%q) on rowMatch = %v, want %v", tc.where, res1, tc.matchRow1)
		}
		if res2 := expr.Eval(rowNonMatch); res2 != tc.matchRow2 {
			t.Errorf("Eval(%q) on rowNonMatch = %v, want %v", tc.where, res2, tc.matchRow2)
		}
	}
}

func TestParseWhere_Errors(t *testing.T) {
	cols := sampleColumns()

	invalidColTests := []string{
		"unknown_col = 1",
		"age > 20 AND non_existent = 'test'",
	}
	for _, q := range invalidColTests {
		_, err := service.ParseWhere(q, cols)
		if err == nil {
			t.Errorf("expected error for invalid column in %q, got nil", q)
		}
		if !service.IsInvalidColumn(err) {
			t.Errorf("expected InvalidColumnError for %q, got %T: %v", q, err, err)
		}
	}

	syntaxErrorTests := []string{
		"age >",
		"status =",
		"(age > 20",
		"name LIKE",
		"status IN (1, 2",
		"status 'active'",
		"age = 'unclosed string",
	}
	for _, q := range syntaxErrorTests {
		_, err := service.ParseWhere(q, cols)
		if err == nil {
			t.Errorf("expected error for syntax error in %q, got nil", q)
		}
		if !service.IsMalformedQuery(err) {
			t.Errorf("expected MalformedQueryError for %q, got %T: %v", q, err, err)
		}
	}
}
