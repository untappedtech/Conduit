package openapi

// OpenAPI represents the root OpenAPI 3.0.3 specification object.
type OpenAPI struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Servers    []Server              `json:"servers,omitempty"`
	Paths      map[string]PathItem   `json:"paths"`
	Components Components            `json:"components"`
	Security   []map[string][]string `json:"security,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	Title          string   `json:"title"`
	Version        string   `json:"version"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Contact describes contact information for the exposed API.
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License describes license information for the exposed API.
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents an API server host and base path.
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// PathItem describes the operations available on a single path.
type PathItem struct {
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Get         *Operation  `json:"get,omitempty"`
	Post        *Operation  `json:"post,omitempty"`
	Put         *Operation  `json:"put,omitempty"`
	Patch       *Operation  `json:"patch,omitempty"`
	Delete      *Operation  `json:"delete,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter describes a single operation parameter (path, query, header).
type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"` // "query", "path", "header"
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// RequestBody describes a single request body.
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Response describes a single response from an API operation.
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType provides schema and examples for a specific media type.
type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

// Schema represents a JSON Schema object for OpenAPI 3.0.3.
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Default     any                `json:"default,omitempty"`
	Enum        []any              `json:"enum,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	Nullable    *bool              `json:"nullable,omitempty"`
	ReadOnly    *bool              `json:"readOnly,omitempty"`
	Example     any                `json:"example,omitempty"`
	Minimum     *int               `json:"minimum,omitempty"`
}

// Components holds reusable objects for different aspects of the OAS.
type Components struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme defines a security scheme that can be used by the operations.
type SecurityScheme struct {
	Type        string `json:"type"` // "http", "apiKey"
	Scheme      string `json:"scheme,omitempty"`
	In          string `json:"in,omitempty"` // "header", "query"
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
