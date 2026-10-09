package service_test

import (
	"fmt"
	"testing"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

func TestWhereCache_ASTAndSQL(t *testing.T) {
	service.ClearWhereCache()
	service.SetCachingEnabled(true)
	defer service.ClearWhereCache()

	cols := []domain.ColumnDef{
		{Name: "id", Type: "INTEGER"},
		{Name: "age", Type: "INTEGER"},
		{Name: "name", Type: "TEXT"},
	}

	whereQuery := "age > 21 AND name = 'Alice'"

	// First call - cache miss
	ast1, err := service.ParseWhere(whereQuery, cols)
	if err != nil {
		t.Fatalf("unexpected error on ParseWhere: %v", err)
	}

	// Second call - should hit AST cache and return the same pointer
	ast2, err := service.ParseWhere(whereQuery, cols)
	if err != nil {
		t.Fatalf("unexpected error on second ParseWhere: %v", err)
	}
	if ast1 != ast2 {
		t.Fatalf("expected identical AST pointer from cache, got %p vs %p", ast1, ast2)
	}

	// Test ParseWhereSQL caching
	sql1, args1, err := service.ParseWhereSQL(whereQuery, cols, service.SimpleDialect{}, 1)
	if err != nil {
		t.Fatalf("unexpected error on ParseWhereSQL: %v", err)
	}

	sql2, args2, err := service.ParseWhereSQL(whereQuery, cols, service.SimpleDialect{}, 1)
	if err != nil {
		t.Fatalf("unexpected error on second ParseWhereSQL: %v", err)
	}

	if sql1 != sql2 {
		t.Fatalf("expected matching SQL: %q vs %q", sql1, sql2)
	}
	if len(args1) != len(args2) {
		t.Fatalf("expected matching args len: %d vs %d", len(args1), len(args2))
	}
	for i := range args1 {
		if fmt.Sprintf("%v", args1[i]) != fmt.Sprintf("%v", args2[i]) {
			t.Errorf("args[%d] mismatch: %v vs %v", i, args1[i], args2[i])
		}
	}

	// Verify mutating returned args does not corrupt cache
	args1[0] = 99999
	_, freshArgs, _ := service.ParseWhereSQL(whereQuery, cols, service.SimpleDialect{}, 1)
	if freshArgs[0] == 99999 {
		t.Fatalf("cache was mutated by caller!")
	}
}

func TestWhereCache_DisableAndClear(t *testing.T) {
	service.ClearWhereCache()
	defer service.ClearWhereCache()

	cols := []domain.ColumnDef{
		{Name: "id", Type: "INTEGER"},
	}

	service.SetCachingEnabled(false)
	ast1, err := service.ParseWhere("id = 1", cols)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	ast2, err := service.ParseWhere("id = 1", cols)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ast1 == ast2 {
		t.Fatalf("expected different pointers when caching disabled")
	}

	service.SetCachingEnabled(true)
	ast3, _ := service.ParseWhere("id = 1", cols)
	ast4, _ := service.ParseWhere("id = 1", cols)
	if ast3 != ast4 {
		t.Fatalf("expected identical pointers when caching enabled")
	}

	service.ClearWhereCache()
	ast5, _ := service.ParseWhere("id = 1", cols)
	if ast4 == ast5 {
		t.Fatalf("expected new pointer after ClearWhereCache")
	}
}
