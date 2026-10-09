package service_test

import (
	"testing"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

func FuzzWhereParser(f *testing.F) {
	// Seed corpus with valid expressions
	f.Add("")
	f.Add("id = 1")
	f.Add("name = 'Alice'")
	f.Add("age > 21")
	f.Add("price <= 99.95")
	f.Add("active = true")
	f.Add("active = false")
	f.Add("deleted_at IS NULL")
	f.Add("deleted_at IS NOT NULL")
	f.Add("status IN ('pending', 'completed')")
	f.Add("status NOT IN ('archived', 'deleted')")
	f.Add("name LIKE '%test%'")
	f.Add("age >= 18 AND active = true")
	f.Add("age < 18 OR name = 'Admin'")
	f.Add("NOT (status = 'banned')")
	f.Add("id = 1 AND (age > 20 OR (active = true AND status = 'ok'))")
	f.Add("name = 'O\\'Reilly'")
	f.Add("id == 5")
	f.Add("id <> 10")
	f.Add("id != 10")

	// Seed corpus with malformed / edge-case expressions
	f.Add("id =")
	f.Add("=")
	f.Add("AND")
	f.Add("OR")
	f.Add("NOT")
	f.Add("(")
	f.Add(")")
	f.Add("()")
	f.Add("(((")
	f.Add(")))")
	f.Add("'unclosed string")
	f.Add("\"unclosed double quote")
	f.Add("'string with trailing backslash\\")
	f.Add("id IN ()")
	f.Add("id IN (1,")
	f.Add("id IN (1, 2")
	f.Add("id IN (1, 2,)")
	f.Add("id IS")
	f.Add("id IS NOT")
	f.Add("id IS 123")
	f.Add("id >=")
	f.Add("id ===")
	f.Add("id NOT IN")
	f.Add("123 = id")
	f.Add("unknown_column = 1")
	f.Add("-")
	f.Add("-.")
	f.Add(".")
	f.Add("1.2.3")
	f.Add("1e10")
	f.Add("\x00")
	f.Add("\xff\xfe")
	f.Add("null = null")
	f.Add("true = false")
	f.Add("name = ''")
	f.Add("status = NULL")

	cols := []domain.ColumnDef{
		{Name: "id", Type: "INTEGER"},
		{Name: "name", Type: "TEXT"},
		{Name: "age", Type: "INTEGER"},
		{Name: "price", Type: "REAL"},
		{Name: "active", Type: "BOOLEAN"},
		{Name: "status", Type: "TEXT"},
		{Name: "deleted_at", Type: "TEXT"},
	}

	dummyRow := map[string]any{
		"id":         int64(1),
		"name":       "Alice",
		"age":        int64(25),
		"price":      49.99,
		"active":     true,
		"status":     "pending",
		"deleted_at": nil,
	}

	f.Fuzz(func(t *testing.T, expr string) {
		// Neither ParseWhere nor subsequent AST evaluation/SQL conversion should panic
		ast, err := service.ParseWhere(expr, cols)
		if err != nil || ast == nil {
			return
		}

		// Ensure ToSQL does not panic
		sql, args := ast.ToSQL(service.SimpleDialect{}, 1)
		_ = sql
		_ = args

		// Ensure ParseWhereSQL does not panic
		sql2, args2, err2 := service.ParseWhereSQL(expr, cols, service.SimpleDialect{}, 1)
		_ = sql2
		_ = args2
		_ = err2

		// Ensure Eval does not panic
		_ = ast.Eval(dummyRow)
	})
}
