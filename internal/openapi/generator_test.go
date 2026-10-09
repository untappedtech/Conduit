package openapi_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/untappedtech/conduit/internal/db/impl"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/openapi"
	"github.com/untappedtech/conduit/internal/service"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestSQLTypeToSchema(t *testing.T) {
	tests := []struct {
		name       string
		col        domain.ColumnDef
		expectType string
		expectFmt  string
		expectNull bool
		expectRO   bool
	}{
		{
			name:       "Integer PK autoincrement",
			col:        domain.ColumnDef{Name: "id", Type: "INTEGER", PK: boolPtr(true), Autoincrement: boolPtr(true)},
			expectType: "integer",
			expectFmt:  "int32",
			expectRO:   true,
		},
		{
			name:       "Bigint",
			col:        domain.ColumnDef{Name: "big_val", Type: "BIGINT"},
			expectType: "integer",
			expectFmt:  "int64",
		},
		{
			name:       "Varchar with length",
			col:        domain.ColumnDef{Name: "name", Type: "VARCHAR(255)", Nullable: boolPtr(true)},
			expectType: "string",
			expectNull: true,
		},
		{
			name:       "Timestamp with time zone",
			col:        domain.ColumnDef{Name: "created_at", Type: "TIMESTAMP WITH TIME ZONE"},
			expectType: "string",
			expectFmt:  "date-time",
		},
		{
			name:       "Numeric / Decimal",
			col:        domain.ColumnDef{Name: "amount", Type: "DECIMAL(10,2)"},
			expectType: "number",
			expectFmt:  "double",
		},
		{
			name:       "Boolean",
			col:        domain.ColumnDef{Name: "active", Type: "BOOLEAN"},
			expectType: "boolean",
		},
		{
			name:       "Blob",
			col:        domain.ColumnDef{Name: "data", Type: "BLOB"},
			expectType: "string",
			expectFmt:  "binary",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := openapi.SQLTypeToSchema(tc.col)
			if s.Type != tc.expectType {
				t.Fatalf("expected type %s, got %s", tc.expectType, s.Type)
			}
			if tc.expectFmt != "" && s.Format != tc.expectFmt {
				t.Fatalf("expected format %s, got %s", tc.expectFmt, s.Format)
			}
			if tc.expectNull && (s.Nullable == nil || !*s.Nullable) {
				t.Fatalf("expected nullable true")
			}
			if tc.expectRO && (s.ReadOnly == nil || !*s.ReadOnly) {
				t.Fatalf("expected readOnly true")
			}
		})
	}
}

func TestGenerator_GenerateJSON(t *testing.T) {
	memDB := impl.NewMemoryDB()
	ctx := context.Background()

	// Create test tables
	err := memDB.CreateTable(ctx, "authors", []domain.ColumnDef{
		{Name: "id", Type: "INTEGER", PK: boolPtr(true), Autoincrement: boolPtr(true)},
		{Name: "name", Type: "TEXT", Nullable: boolPtr(false)},
		{Name: "bio", Type: "TEXT", Nullable: boolPtr(true)},
	})
	if err != nil {
		t.Fatalf("failed to create authors table: %v", err)
	}

	err = memDB.CreateTable(ctx, "books", []domain.ColumnDef{
		{Name: "id", Type: "INTEGER", PK: boolPtr(true), Autoincrement: boolPtr(true)},
		{Name: "title", Type: "VARCHAR(200)", Nullable: boolPtr(false)},
		{Name: "price", Type: "DECIMAL(8,2)", Nullable: boolPtr(false)},
		{Name: "author_id", Type: "INTEGER", Nullable: boolPtr(true)},
	})
	if err != nil {
		t.Fatalf("failed to create books table: %v", err)
	}

	serverConfig := &domain.ServerConfig{}
	serverConfig.Server.BasePath = "/v1"
	serverConfig.Server.DefaultLimit = 25
	serverConfig.Policy.PublicReads = true
	serverConfig.Policy.PublicWrites = true

	apiService := service.NewAPIService(memDB, serverConfig)
	generator := openapi.NewGenerator(apiService, serverConfig)

	data, err := generator.GenerateJSON(ctx)
	if err != nil {
		t.Fatalf("failed to generate openapi json: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid json generated: %v", err)
	}

	// Verify OpenAPI version
	if parsed["openapi"] != "3.0.3" {
		t.Fatalf("expected openapi 3.0.3, got %v", parsed["openapi"])
	}

	paths, ok := parsed["paths"].(map[string]any)
	if !ok {
		t.Fatalf("expected paths object")
	}

	// Verify Schema paths exist
	if _, ok := paths["/schema"]; !ok {
		t.Fatalf("missing /schema path")
	}
	if _, ok := paths["/schema/{table}"]; !ok {
		t.Fatalf("missing /schema/{table} path")
	}

	// Verify Table paths exist
	if _, ok := paths["/authors"]; !ok {
		t.Fatalf("missing /authors path")
	}
	if _, ok := paths["/authors/{id}"]; !ok {
		t.Fatalf("missing /authors/{id} path")
	}
	if _, ok := paths["/books"]; !ok {
		t.Fatalf("missing /books path")
	}
	if _, ok := paths["/books/{id}"]; !ok {
		t.Fatalf("missing /books/{id} path")
	}

	// Inspect /books GET parameters
	booksPath := paths["/books"].(map[string]any)
	getOp := booksPath["get"].(map[string]any)
	params := getOp["parameters"].([]any)

	foundWhere := false
	foundOrder := false
	foundLimit := false
	for _, p := range params {
		paramMap := p.(map[string]any)
		name := paramMap["name"].(string)
		if name == "where" {
			foundWhere = true
		}
		if name == "order" {
			foundOrder = true
		}
		if name == "limit" {
			foundLimit = true
			schema := paramMap["schema"].(map[string]any)
			if schema["default"].(float64) != 25 {
				t.Fatalf("expected limit default 25, got %v", schema["default"])
			}
		}
	}

	if !foundWhere || !foundOrder || !foundLimit {
		t.Fatalf("expected where, order, limit query parameters (where=%v, order=%v, limit=%v)", foundWhere, foundOrder, foundLimit)
	}

	// Check Components.Schemas
	components := parsed["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	if _, ok := schemas["books"]; !ok {
		t.Fatalf("missing books schema")
	}
	if _, ok := schemas["booksInput"]; !ok {
		t.Fatalf("missing booksInput schema")
	}

	// Check booksInput required fields: title, price (id is autoincrement, author_id is nullable)
	booksInput := schemas["booksInput"].(map[string]any)
	reqFields, ok := booksInput["required"].([]any)
	if !ok {
		t.Fatalf("expected required array on booksInput")
	}
	if len(reqFields) != 2 {
		t.Fatalf("expected 2 required fields on booksInput, got %d: %v", len(reqFields), reqFields)
	}
}

func TestGenerator_Invalidate(t *testing.T) {
	memDB := impl.NewMemoryDB()
	ctx := context.Background()

	serverConfig := &domain.ServerConfig{}
	apiService := service.NewAPIService(memDB, serverConfig)
	generator := openapi.NewGenerator(apiService, serverConfig)

	// First generation: no tables
	d1, err := generator.GenerateJSON(ctx)
	if err != nil {
		t.Fatalf("gen 1 failed: %v", err)
	}

	// Create table
	_ = memDB.CreateTable(ctx, "notes", []domain.ColumnDef{
		{Name: "id", Type: "INTEGER", PK: boolPtr(true)},
		{Name: "body", Type: "TEXT"},
	})

	// Without invalidate, it should return cached json (still no notes)
	d2, _ := generator.GenerateJSON(ctx)
	if string(d1) != string(d2) {
		t.Fatalf("expected cached spec to match")
	}

	// Invalidate cache
	generator.Invalidate()

	d3, err := generator.GenerateJSON(ctx)
	if err != nil {
		t.Fatalf("gen 3 failed: %v", err)
	}
	if string(d1) == string(d3) {
		t.Fatalf("expected invalidated spec to be different after creating table")
	}
}

func TestDocsHTML(t *testing.T) {
	html := openapi.DocsHTML("/v1/openapi.json", "Test Title")
	htmlStr := string(html)
	if len(htmlStr) == 0 {
		t.Fatalf("empty html")
	}
	if !testing.Short() {
		if !contains(htmlStr, "Test Title") {
			t.Fatalf("expected title in html")
		}
		if !contains(htmlStr, "/v1/openapi.json") {
			t.Fatalf("expected spec url in html")
		}
		if !contains(htmlStr, "@scalar/api-reference") {
			t.Fatalf("expected scalar script in html")
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && index(s, substr) >= 0))
}

func index(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
