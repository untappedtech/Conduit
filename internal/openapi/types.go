package openapi

import (
	"strings"

	"github.com/untappedtech/conduit/internal/domain"
)

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

// normalizeSQLType strips length/precision qualifiers (e.g. VARCHAR(255) -> VARCHAR).
func normalizeSQLType(rawType string) string {
	t := strings.ToUpper(strings.TrimSpace(rawType))
	if idx := strings.Index(t, "("); idx != -1 {
		t = strings.TrimSpace(t[:idx])
	}
	return t
}

// SQLTypeToSchema converts a domain.ColumnDef into an OpenAPI Schema representation.
func SQLTypeToSchema(col domain.ColumnDef) *Schema {
	norm := normalizeSQLType(col.Type)
	s := &Schema{}

	switch norm {
	case "INT", "INTEGER", "INT4", "MEDIUMINT":
		s.Type = "integer"
		s.Format = "int32"
	case "BIGINT", "INT8", "BIGSERIAL":
		s.Type = "integer"
		s.Format = "int64"
	case "SMALLINT", "INT2", "TINYINT", "SMALLSERIAL":
		s.Type = "integer"
		s.Format = "int32"
	case "SERIAL":
		s.Type = "integer"
		s.Format = "int32"
	case "REAL", "FLOAT", "FLOAT4":
		s.Type = "number"
		s.Format = "float"
	case "DOUBLE", "DOUBLE PRECISION", "FLOAT8", "NUMERIC", "DECIMAL", "NUMBER":
		s.Type = "number"
		s.Format = "double"
	case "BOOL", "BOOLEAN", "BIT":
		s.Type = "boolean"
	case "DATE":
		s.Type = "string"
		s.Format = "date"
	case "TIME", "TIMETZ", "TIME WITHOUT TIME ZONE", "TIME WITH TIME ZONE":
		s.Type = "string"
		s.Format = "time"
	case "TIMESTAMP", "TIMESTAMPTZ", "DATETIME", "TIMESTAMP WITHOUT TIME ZONE", "TIMESTAMP WITH TIME ZONE":
		s.Type = "string"
		s.Format = "date-time"
	case "UUID":
		s.Type = "string"
		s.Format = "uuid"
	case "BLOB", "BYTEA", "BINARY", "VARBINARY", "IMAGE", "RAW":
		s.Type = "string"
		s.Format = "binary"
	case "JSON", "JSONB":
		s.Type = "object"
	default:
		s.Type = "string"
	}

	if col.Nullable != nil && *col.Nullable {
		s.Nullable = boolPtr(true)
	}

	if col.Autoincrement != nil && *col.Autoincrement {
		s.ReadOnly = boolPtr(true)
	}

	if col.Default != nil {
		s.Default = col.DefaultValue()
	}

	return s
}

// BuildTableSchemas generates both the response record schema and request input schema for a table.
func BuildTableSchemas(tableName string, columns []domain.ColumnDef) (recordSchema *Schema, inputSchema *Schema) {
	recordProperties := make(map[string]*Schema)
	inputProperties := make(map[string]*Schema)
	var requiredFields []string

	for _, col := range columns {
		colSchema := SQLTypeToSchema(col)
		recordProperties[col.Name] = colSchema

		isAuto := col.Autoincrement != nil && *col.Autoincrement
		if !isAuto {
			inputColSchema := SQLTypeToSchema(col)
			inputColSchema.ReadOnly = nil // Auto-increment is not read-only in input payload
			inputProperties[col.Name] = inputColSchema

			isNullable := col.Nullable != nil && *col.Nullable
			hasDefault := col.Default != nil
			if !isNullable && !hasDefault {
				requiredFields = append(requiredFields, col.Name)
			}
		}
	}

	recordSchema = &Schema{
		Type:       "object",
		Properties: recordProperties,
	}

	inputSchema = &Schema{
		Type:       "object",
		Properties: inputProperties,
	}
	if len(requiredFields) > 0 {
		inputSchema.Required = requiredFields
	}

	return recordSchema, inputSchema
}
