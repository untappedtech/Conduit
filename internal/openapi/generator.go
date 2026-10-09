package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

// Generator builds an OpenAPI 3.0.3 specification by inspecting the active database schema.
type Generator struct {
	apiService   *service.APIService
	serverConfig *domain.ServerConfig
	mutex        sync.RWMutex
	cachedJSON   []byte
}

// NewGenerator creates a new OpenAPI Generator instance.
func NewGenerator(apiService *service.APIService, serverConfig *domain.ServerConfig) *Generator {
	return &Generator{
		apiService:   apiService,
		serverConfig: serverConfig,
	}
}

// Invalidate clears the cached OpenAPI JSON document, forcing regeneration on the next request.
func (g *Generator) Invalidate() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.cachedJSON = nil
}

// GenerateJSON returns the indented JSON representation of the OpenAPI specification.
func (g *Generator) GenerateJSON(ctx context.Context) ([]byte, error) {
	g.mutex.RLock()
	if g.cachedJSON != nil {
		cached := g.cachedJSON
		g.mutex.RUnlock()
		return cached, nil
	}
	g.mutex.RUnlock()

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Double-check after acquiring write lock
	if g.cachedJSON != nil {
		return g.cachedJSON, nil
	}

	spec, err := g.Generate(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openapi spec to JSON: %w", err)
	}

	g.cachedJSON = data
	return data, nil
}

// Generate constructs the complete OpenAPI 3.0.3 specification.
func (g *Generator) Generate(ctx context.Context) (*OpenAPI, error) {
	basePath := "/v1"
	defaultLimit := 50

	title := "Conduit API"
	version := "1.0.0"
	description := "Multi-Format, Database-Agnostic REST Engine API documentation automatically generated from relational database schema."
	var termsOfService string
	var contact *Contact
	var license *License
	var servers []Server

	if g.serverConfig != nil {
		if g.serverConfig.Server.BasePath != "" {
			bp := strings.TrimSpace(g.serverConfig.Server.BasePath)
			if !strings.HasPrefix(bp, "/") {
				bp = "/" + bp
			}
			basePath = strings.TrimRight(bp, "/")
			if basePath == "" {
				basePath = "/"
			}
		}
		if g.serverConfig.Server.DefaultLimit > 0 {
			defaultLimit = g.serverConfig.Server.DefaultLimit
		}

		if g.serverConfig.OpenAPI.Title != "" {
			title = g.serverConfig.OpenAPI.Title
		}
		if g.serverConfig.OpenAPI.Version != "" {
			version = g.serverConfig.OpenAPI.Version
		}
		if g.serverConfig.OpenAPI.Description != "" {
			description = g.serverConfig.OpenAPI.Description
		}
		if g.serverConfig.OpenAPI.TermsOfService != "" {
			termsOfService = g.serverConfig.OpenAPI.TermsOfService
		}
		if g.serverConfig.OpenAPI.Contact != nil {
			if g.serverConfig.OpenAPI.Contact.Name != "" || g.serverConfig.OpenAPI.Contact.Email != "" || g.serverConfig.OpenAPI.Contact.URL != "" {
				contact = &Contact{
					Name:  g.serverConfig.OpenAPI.Contact.Name,
					Email: g.serverConfig.OpenAPI.Contact.Email,
					URL:   g.serverConfig.OpenAPI.Contact.URL,
				}
			}
		}
		if g.serverConfig.OpenAPI.License != nil && g.serverConfig.OpenAPI.License.Name != "" {
			license = &License{
				Name: g.serverConfig.OpenAPI.License.Name,
				URL:  g.serverConfig.OpenAPI.License.URL,
			}
		}
		if len(g.serverConfig.OpenAPI.Servers) > 0 {
			for _, s := range g.serverConfig.OpenAPI.Servers {
				if s.URL != "" {
					servers = append(servers, Server{
						URL:         s.URL,
						Description: s.Description,
					})
				}
			}
		}
	}

	if len(servers) == 0 {
		servers = []Server{
			{
				URL:         basePath,
				Description: "Conduit API Server",
			},
		}
	}

	spec := &OpenAPI{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:          title,
			Version:        version,
			Description:    description,
			TermsOfService: termsOfService,
			Contact:        contact,
			License:        license,
		},
		Servers: servers,
		Paths:   make(map[string]PathItem),
		Components: Components{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	// Register shared error schemas
	spec.Components.Schemas["ErrorDetail"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"title":             {Type: "string", Description: "Error title"},
			"message":           {Type: "string", Description: "Detailed error message"},
			"status":            {Type: "integer", Description: "HTTP status code"},
			"documentation_url": {Type: "string", Description: "Link to relevant error documentation"},
			"details": {
				Type:        "array",
				Items:       &Schema{Type: "string"},
				Description: "Additional error context or stack hints",
			},
		},
		Required: []string{"title", "message", "status"},
	}

	spec.Components.Schemas["ErrorResponse"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"error": {
				Ref: "#/components/schemas/ErrorDetail",
			},
		},
		Required: []string{"error"},
	}

	// Register shared schema management schemas
	spec.Components.Schemas["ColumnDef"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name":          {Type: "string", Description: "Column identifier"},
			"type":          {Type: "string", Description: "Database SQL data type"},
			"nullable":      {Type: "boolean", Description: "Accepts NULL values"},
			"unique":        {Type: "boolean", Description: "Column has UNIQUE constraint"},
			"default":       {Description: "Default value for column"},
			"pk":            {Type: "boolean", Description: "Column is primary key"},
			"autoincrement": {Type: "boolean", Description: "Auto-incrementing identity"},
			"cid":           {Type: "integer", Description: "Column order index"},
		},
		Required: []string{"name", "type"},
	}

	spec.Components.Schemas["CreateTableRequest"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"columns": {
				Type:  "array",
				Items: &Schema{Ref: "#/components/schemas/ColumnDef"},
			},
		},
		Required: []string{"columns"},
	}

	// Register Security Schemes
	spec.Components.SecuritySchemes["BearerAuth"] = SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: "Bearer token authentication passed via Authorization: Bearer <token>",
	}
	spec.Components.SecuritySchemes["ApiKeyHeader"] = SecurityScheme{
		Type:        "apiKey",
		In:          "header",
		Name:        "X-API-Key",
		Description: "API key passed via X-API-Key header",
	}
	spec.Components.SecuritySchemes["ApiKeyQuery"] = SecurityScheme{
		Type:        "apiKey",
		In:          "query",
		Name:        "api_key",
		Description: "API key passed via api_key query parameter",
	}

	if g.serverConfig != nil && g.serverConfig.Policy.PublicReads && g.serverConfig.Policy.PublicWrites {
		spec.Security = []map[string][]string{
			{},
			{"BearerAuth": {}},
			{"ApiKeyHeader": {}},
			{"ApiKeyQuery": {}},
		}
	} else {
		spec.Security = []map[string][]string{
			{"BearerAuth": {}},
			{"ApiKeyHeader": {}},
			{"ApiKeyQuery": {}},
		}
	}

	// Add Schema Management routes
	g.registerSchemaManagementPaths(spec)

	if g.apiService == nil {
		return spec, nil
	}

	tables, err := g.apiService.ListTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list database tables: %w", err)
	}

	for _, tableName := range tables {
		columns, err := g.apiService.GetSchema(ctx, tableName)
		if err != nil {
			continue
		}

		recordSchema, inputSchema := BuildTableSchemas(tableName, columns)
		spec.Components.Schemas[tableName] = recordSchema
		spec.Components.Schemas[tableName+"Input"] = inputSchema

		g.registerTablePaths(spec, tableName, columns, defaultLimit)
	}

	return spec, nil
}

func (g *Generator) registerSchemaManagementPaths(spec *OpenAPI) {
	spec.Paths["/schema"] = PathItem{
		Get: &Operation{
			Tags:        []string{"Schema"},
			Summary:     "List all tables",
			Description: "Returns a list of all table names currently present in the database.",
			Responses: map[string]Response{
				"200": {
					Description: "List of table names",
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{
								Type:  "array",
								Items: &Schema{Type: "string"},
							},
						},
					},
				},
				"500": {
					Description: "Internal Server Error",
					Content: map[string]MediaType{
						"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
					},
				},
			},
		},
	}

	spec.Paths["/schema/{table}"] = PathItem{
		Parameters: []Parameter{
			{
				Name:        "table",
				In:          "path",
				Required:    true,
				Description: "Target database table name",
				Schema:      &Schema{Type: "string"},
			},
		},
		Get: &Operation{
			Tags:        []string{"Schema"},
			Summary:     "Get table schema",
			Description: "Returns column definitions and constraints for the specified table.",
			Responses: map[string]Response{
				"200": {
					Description: "Column definitions",
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{
								Type:  "array",
								Items: &Schema{Ref: "#/components/schemas/ColumnDef"},
							},
						},
					},
				},
				"404": {
					Description: "Table not found",
					Content: map[string]MediaType{
						"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
					},
				},
			},
		},
		Post: &Operation{
			Tags:        []string{"Schema"},
			Summary:     "Create table",
			Description: "Creates a new table with the provided column definitions.",
			RequestBody: &RequestBody{
				Required:    true,
				Description: "Table schema definition containing column list",
				Content: map[string]MediaType{
					"application/json":   {Schema: &Schema{Ref: "#/components/schemas/CreateTableRequest"}},
					"application/x-yaml": {Schema: &Schema{Ref: "#/components/schemas/CreateTableRequest"}},
					"application/toml":   {Schema: &Schema{Ref: "#/components/schemas/CreateTableRequest"}},
				},
			},
			Responses: map[string]Response{
				"201": {
					Description: "Table created successfully",
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{
								Type:  "array",
								Items: &Schema{Ref: "#/components/schemas/ColumnDef"},
							},
						},
					},
				},
				"400": {
					Description: "Invalid column payload or syntax error",
					Content: map[string]MediaType{
						"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
					},
				},
			},
		},
		Delete: &Operation{
			Tags:        []string{"Schema"},
			Summary:     "Drop table",
			Description: "Permanently drops the specified table and its contents.",
			Responses: map[string]Response{
				"204": {
					Description: "Table dropped successfully",
				},
				"500": {
					Description: "Failed to drop table",
					Content: map[string]MediaType{
						"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
					},
				},
			},
		},
	}
}

func (g *Generator) registerTablePaths(spec *OpenAPI, tableName string, columns []domain.ColumnDef, defaultLimit int) {
	recordRef := "#/components/schemas/" + tableName
	inputRef := "#/components/schemas/" + tableName + "Input"

	// Multi-format response definitions for record collections
	listResponses := map[string]Response{
		"200": {
			Description: fmt.Sprintf("List of %s records", tableName),
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{Type: "array", Items: &Schema{Ref: recordRef}},
				},
				"application/xml": {
					Schema: &Schema{Type: "array", Items: &Schema{Ref: recordRef}},
				},
				"application/x-yaml": {
					Schema: &Schema{Type: "array", Items: &Schema{Ref: recordRef}},
				},
				"application/toml": {
					Schema: &Schema{Type: "array", Items: &Schema{Ref: recordRef}},
				},
				"text/csv": {
					Schema: &Schema{Type: "string"},
				},
				"application/x-ndjson": {
					Schema: &Schema{Type: "string"},
				},
				"application/cbor": {
					Schema: &Schema{Type: "string", Format: "binary"},
				},
			},
		},
		"400": {
			Description: "Bad Request (e.g. invalid sort column or query syntax)",
			Content: map[string]MediaType{
				"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
			},
		},
		"422": {
			Description: "Unprocessable Entity (e.g. malformed where filter expression)",
			Content: map[string]MediaType{
				"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
			},
		},
	}

	// Multi-format request body for insert / update
	recordRequestBody := &RequestBody{
		Required:    true,
		Description: fmt.Sprintf("%s record payload", tableName),
		Content: map[string]MediaType{
			"application/json":   {Schema: &Schema{Ref: inputRef}},
			"application/x-yaml": {Schema: &Schema{Ref: inputRef}},
			"application/xml":    {Schema: &Schema{Ref: inputRef}},
			"application/toml":   {Schema: &Schema{Ref: inputRef}},
		},
	}

	// Single record responses
	singleRecordResponses := map[string]Response{
		"200": {
			Description: fmt.Sprintf("Single %s record", tableName),
			Content: map[string]MediaType{
				"application/json":   {Schema: &Schema{Ref: recordRef}},
				"application/xml":    {Schema: &Schema{Ref: recordRef}},
				"application/x-yaml": {Schema: &Schema{Ref: recordRef}},
				"application/toml":   {Schema: &Schema{Ref: recordRef}},
				"application/cbor":   {Schema: &Schema{Type: "string", Format: "binary"}},
			},
		},
		"404": {
			Description: "Record not found",
			Content: map[string]MediaType{
				"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
			},
		},
	}

	// Path: /{table}
	tablePath := "/" + tableName
	spec.Paths[tablePath] = PathItem{
		Get: &Operation{
			Tags:        []string{tableName},
			Summary:     fmt.Sprintf("List %s records", tableName),
			Description: fmt.Sprintf("Retrieves a list of %s records with optional filtering, sorting, and pagination.", tableName),
			Parameters: []Parameter{
				{
					Name:        "limit",
					In:          "query",
					Description: "Maximum number of rows to return (set to 0 for unlimited mode).",
					Required:    false,
					Schema:      &Schema{Type: "integer", Default: defaultLimit, Minimum: intPtr(0)},
				},
				{
					Name:        "offset",
					In:          "query",
					Description: "Number of rows to skip before collecting results.",
					Required:    false,
					Schema:      &Schema{Type: "integer", Default: 0, Minimum: intPtr(0)},
				},
				{
					Name:        "order",
					In:          "query",
					Description: "Column sort expression (e.g. col, col:asc, col:desc).",
					Required:    false,
					Schema:      &Schema{Type: "string", Example: "id:asc"},
				},
				{
					Name:        "where",
					In:          "query",
					Description: "SQL-like filter expression evaluated via Conduit AST. Supports operators: =, !=, <, >, <=, >=, LIKE, IN, NOT IN, IS NULL, IS NOT NULL, AND, OR, NOT with grouping parentheses. Example: (status = 'active' OR role = 'admin') AND age >= 21",
					Required:    false,
					Schema:      &Schema{Type: "string", Example: "status = 'active'"},
				},
				{
					Name:        "format",
					In:          "query",
					Description: "Output serialization format.",
					Required:    false,
					Schema: &Schema{
						Type:    "string",
						Enum:    []any{"json", "ndjson", "yaml", "toml", "xml", "csv", "cbor"},
						Default: "json",
					},
				},
			},
			Responses: listResponses,
		},
		Post: &Operation{
			Tags:        []string{tableName},
			Summary:     fmt.Sprintf("Create %s record", tableName),
			Description: fmt.Sprintf("Inserts a new record into %s.", tableName),
			RequestBody: recordRequestBody,
			Responses: map[string]Response{
				"201": {
					Description: "Record successfully created",
					Content: map[string]MediaType{
						"application/json":   {Schema: &Schema{Ref: recordRef}},
						"application/x-yaml": {Schema: &Schema{Ref: recordRef}},
						"application/xml":    {Schema: &Schema{Ref: recordRef}},
						"application/toml":   {Schema: &Schema{Ref: recordRef}},
					},
				},
				"400": {
					Description: "Invalid payload or column constraint violation",
					Content: map[string]MediaType{
						"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
					},
				},
			},
		},
	}

	// Locate single Primary Key column for /{table}/{id}
	var pkCol *domain.ColumnDef
	pkCount := 0
	for i := range columns {
		if columns[i].PK != nil && *columns[i].PK {
			pkCol = &columns[i]
			pkCount++
		}
	}

	// If table has exactly one PK column, expose /{table}/{id}
	if pkCount == 1 && pkCol != nil {
		idPath := fmt.Sprintf("/%s/{id}", tableName)
		idParamSchema := SQLTypeToSchema(*pkCol)
		idParamSchema.ReadOnly = nil
		idParamSchema.Nullable = nil

		spec.Paths[idPath] = PathItem{
			Parameters: []Parameter{
				{
					Name:        "id",
					In:          "path",
					Required:    true,
					Description: fmt.Sprintf("Primary key identifier (%s) for the record", pkCol.Name),
					Schema:      idParamSchema,
				},
			},
			Get: &Operation{
				Tags:        []string{tableName},
				Summary:     fmt.Sprintf("Get %s record by ID", tableName),
				Description: fmt.Sprintf("Retrieves a single record from %s by primary key.", tableName),
				Responses:   singleRecordResponses,
			},
			Put: &Operation{
				Tags:        []string{tableName},
				Summary:     fmt.Sprintf("Replace %s record by ID", tableName),
				Description: fmt.Sprintf("Replaces an existing record in %s.", tableName),
				RequestBody: recordRequestBody,
				Responses: map[string]Response{
					"200": {
						Description: "Record successfully updated",
						Content: map[string]MediaType{
							"application/json":   {Schema: &Schema{Ref: recordRef}},
							"application/x-yaml": {Schema: &Schema{Ref: recordRef}},
							"application/xml":    {Schema: &Schema{Ref: recordRef}},
							"application/toml":   {Schema: &Schema{Ref: recordRef}},
						},
					},
					"400": {
						Description: "Bad Request",
						Content: map[string]MediaType{
							"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
						},
					},
					"404": {
						Description: "Record not found",
						Content: map[string]MediaType{
							"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
						},
					},
				},
			},
			Patch: &Operation{
				Tags:        []string{tableName},
				Summary:     fmt.Sprintf("Update %s record by ID", tableName),
				Description: fmt.Sprintf("Updates fields of an existing record in %s.", tableName),
				RequestBody: recordRequestBody,
				Responses: map[string]Response{
					"200": {
						Description: "Record successfully updated",
						Content: map[string]MediaType{
							"application/json":   {Schema: &Schema{Ref: recordRef}},
							"application/x-yaml": {Schema: &Schema{Ref: recordRef}},
							"application/xml":    {Schema: &Schema{Ref: recordRef}},
							"application/toml":   {Schema: &Schema{Ref: recordRef}},
						},
					},
					"400": {
						Description: "Bad Request",
						Content: map[string]MediaType{
							"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
						},
					},
					"404": {
						Description: "Record not found",
						Content: map[string]MediaType{
							"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
						},
					},
				},
			},
			Delete: &Operation{
				Tags:        []string{tableName},
				Summary:     fmt.Sprintf("Delete %s record by ID", tableName),
				Description: fmt.Sprintf("Deletes a record from %s by primary key.", tableName),
				Responses: map[string]Response{
					"204": {
						Description: "Record successfully deleted",
					},
					"404": {
						Description: "Record not found",
						Content: map[string]MediaType{
							"application/json": {Schema: &Schema{Ref: "#/components/schemas/ErrorResponse"}},
						},
					},
				},
			},
		}
	}
}
