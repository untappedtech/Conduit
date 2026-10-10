package service_test

import (
	"strings"
	"testing"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

func testColumns() []domain.ColumnDef {
	return []domain.ColumnDef{
		{Name: "id", Type: "INTEGER"},
		{Name: "name", Type: "TEXT"},
		{Name: "score", Type: "INTEGER"},
		{Name: "status", Type: "TEXT"},
	}
}

func TestParseWhere_NestingDepthWithinLimit(t *testing.T) {
	service.ClearWhereCache()
	cols := testColumns()

	tests := []struct {
		name string
		expr string
	}{
		{"Depth 0 (no parens)", "id = 1 AND score > 10"},
		{"Depth 1", "(id = 1)"},
		{"Depth 2", "((id = 1))"},
		{"Depth 3", "(((id = 1)))"},
		{"Depth 4", "((((id = 1))))"},
		{"Depth 5 (default limit)", "(((((id = 1)))))"},
		{"Sibling parens (max depth 2)", "((id = 1) AND (score > 10))"},
		{"Complex within limit", "((status = 'active' OR status = 'pending') AND (score >= 100))"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := service.ParseWhere(tc.expr, cols)
			if err != nil {
				t.Fatalf("ParseWhere(%q) failed unexpectedly: %v", tc.expr, err)
			}
			if ast == nil {
				t.Fatalf("ParseWhere(%q) returned nil AST", tc.expr)
			}
		})
	}
}

func TestParseWhere_NestingDepthExceedsDefaultLimit(t *testing.T) {
	service.ClearWhereCache()
	cols := testColumns()

	// Default limit is 5, so 6 nested parens must fail
	deepExpr := "((((((id = 1))))))"

	ast, err := service.ParseWhere(deepExpr, cols)
	if err == nil {
		t.Fatalf("expected ParseWhere to fail for depth 6 with default limit 5, got AST: %v", ast)
	}

	if !service.IsQueryTooComplex(err) {
		t.Fatalf("expected QueryTooComplexError, got %T: %v", err, err)
	}

	complexErr, ok := err.(*service.QueryTooComplexError)
	if !ok {
		t.Fatalf("failed to cast to *QueryTooComplexError")
	}

	if complexErr.Depth != 6 {
		t.Errorf("expected Depth=6, got %d", complexErr.Depth)
	}
	if complexErr.MaxDepth != 5 {
		t.Errorf("expected MaxDepth=5, got %d", complexErr.MaxDepth)
	}

	expectedMsg := "query too complex: nesting depth 6 exceeds maximum allowed depth of 5"
	if complexErr.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, complexErr.Error())
	}
}

func TestParseWhere_CustomMaxQueryDepth(t *testing.T) {
	service.ClearWhereCache()
	cols := testColumns()

	t.Run("MaxQueryDepth = 2", func(t *testing.T) {
		// Depth 2 should succeed
		expr2 := "((id = 1))"
		if _, err := service.ParseWhere(expr2, cols, 2); err != nil {
			t.Fatalf("expected depth 2 to pass with limit 2: %v", err)
		}

		// Depth 3 should fail
		expr3 := "(((id = 1)))"
		_, err := service.ParseWhere(expr3, cols, 2)
		if err == nil {
			t.Fatalf("expected depth 3 to fail with limit 2")
		}
		if !service.IsQueryTooComplex(err) {
			t.Fatalf("expected QueryTooComplexError, got %T: %v", err, err)
		}
	})

	t.Run("MaxQueryDepth = 8", func(t *testing.T) {
		// Depth 6 should succeed with limit 8
		expr6 := "((((((id = 1))))))"
		if _, err := service.ParseWhere(expr6, cols, 8); err != nil {
			t.Fatalf("expected depth 6 to pass with limit 8: %v", err)
		}

		// Depth 9 should fail with limit 8
		expr9 := strings.Repeat("(", 9) + "id = 1" + strings.Repeat(")", 9)
		_, err := service.ParseWhere(expr9, cols, 8)
		if err == nil {
			t.Fatalf("expected depth 9 to fail with limit 8")
		}
		if !service.IsQueryTooComplex(err) {
			t.Fatalf("expected QueryTooComplexError, got %T: %v", err, err)
		}
	})
}

func TestParseWhereSQL_ComplexityBounding(t *testing.T) {
	service.ClearWhereCache()
	cols := testColumns()

	// Default limit 5: 6 levels of parentheses must fail in ParseWhereSQL
	deepExpr := "((((((id = 1))))))"
	_, _, err := service.ParseWhereSQL(deepExpr, cols, service.SimpleDialect{}, 1)
	if err == nil {
		t.Fatalf("expected ParseWhereSQL to fail for depth 6 with default limit 5")
	}
	if !service.IsQueryTooComplex(err) {
		t.Fatalf("expected QueryTooComplexError from ParseWhereSQL, got: %v", err)
	}

	// With explicit custom depth 10: depth 6 must succeed
	sql, args, err := service.ParseWhereSQL(deepExpr, cols, service.SimpleDialect{}, 1, 10)
	if err != nil {
		t.Fatalf("expected ParseWhereSQL with limit 10 to succeed for depth 6: %v", err)
	}
	if sql == "" || len(args) != 1 {
		t.Fatalf("unexpected sql or args: %s, %v", sql, args)
	}
}

