package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/errors"
	"github.com/untappedtech/conduit/internal/openapi"
	"github.com/untappedtech/conduit/internal/service"
)

type APIHandler struct {
	apiService       *service.APIService
	serverConfig     *domain.ServerConfig
	responseEncoder  *ResponseEncoder
	basePath         string
	openAPIGenerator *openapi.Generator
}

func normalizeBasePath(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return "/v1/"
	}
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}
	return base
}

func NewAPIHandler(apiService *service.APIService, serverConfig *domain.ServerConfig, responseEncoder *ResponseEncoder) *APIHandler {
	basePath := "/v1/"
	if serverConfig != nil && serverConfig.Server.BasePath != "" {
		basePath = serverConfig.Server.BasePath
	}
	return &APIHandler{
		apiService:       apiService,
		serverConfig:     serverConfig,
		responseEncoder:  responseEncoder,
		basePath:         normalizeBasePath(basePath),
		openAPIGenerator: openapi.NewGenerator(apiService, serverConfig),
	}
}

func (handler *APIHandler) BasePath() string {
	return handler.basePath
}

func (handler *APIHandler) OpenAPIGenerator() *openapi.Generator {
	return handler.openAPIGenerator
}

func (handler *APIHandler) RegisterRoutes(serveMux *http.ServeMux, customBasePath ...string) {
	if len(customBasePath) > 0 && customBasePath[0] != "" {
		handler.basePath = normalizeBasePath(customBasePath[0])
	}
	base := handler.basePath
	schemaPath := strings.TrimRight(base, "/") + "/schema"

	serveMux.HandleFunc(schemaPath, handler.handleSchema)
	serveMux.HandleFunc(schemaPath+"/", handler.handleSchema)

	openAPIPath := strings.TrimRight(base, "/") + "/openapi.json"
	docsPath := strings.TrimRight(base, "/") + "/docs"

	serveMux.HandleFunc(openAPIPath, handler.handleOpenAPI)
	serveMux.HandleFunc(docsPath, handler.handleDocsUI)
	serveMux.HandleFunc(docsPath+"/", handler.handleDocsUI)

	if base != "/" {
		serveMux.HandleFunc("/openapi.json", handler.handleOpenAPI)
		serveMux.HandleFunc("/docs", handler.handleDocsUI)
		serveMux.HandleFunc("/docs/", handler.handleDocsUI)
	}

	serveMux.HandleFunc(base, handler.handleCRUD)
}

func (handler *APIHandler) handleSchema(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	schemaPath := strings.TrimRight(handler.basePath, "/") + "/schema"
	requestPath := strings.TrimPrefix(r.URL.Path, schemaPath)
	tableName := strings.Trim(requestPath, "/")

	// GET /v1/schema → list tables
	if tableName == "" && r.Method == http.MethodGet {
		tables, err := handler.apiService.ListTables(r.Context())
		if err != nil {
			log.Printf("[ERROR] Failed to list tables: %v", err)
			handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err))
			return
		}
		handler.responseEncoder.EncodeResponse(w, r, http.StatusOK, tables, domain.FormatJSON, "tables", nil)
		return
	}

	// Missing table name
	if tableName == "" {
		handler.responseEncoder.EncodeError(w, r, errors.ErrBadRequest.WithDetails("No table name specified"))
		return
	}

	switch r.Method {

	case http.MethodGet:
		columns, err := handler.apiService.GetSchema(r.Context(), tableName)
		if err != nil {
			handler.responseEncoder.EncodeError(w, r, errors.ErrNotFound.With(err, "Table: "+tableName))
			return
		}
		handler.responseEncoder.EncodeResponse(w, r, http.StatusOK, columns, domain.FormatJSON, "columns", nil)

	case http.MethodPost:
		var payload struct {
			Columns []domain.ColumnDef `json:"columns" yaml:"columns" xml:"column" toml:"columns"`
		}
		format, decodeErr := DecodeInputPayload(r, &payload)
		if decodeErr != nil || len(payload.Columns) == 0 {
			handler.responseEncoder.EncodeError(w, r, errors.ErrBadRequest.With(decodeErr, "Payload must include a 'columns' array"))
			return
		}

		if err := handler.apiService.CreateTable(r.Context(), tableName, payload.Columns); err != nil {
			log.Printf("[ERROR] Failed to create table %s: %v", tableName, err)
			handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err, "Table: "+tableName))
			return
		}

		handler.openAPIGenerator.Invalidate()
		log.Printf("[INFO] Table successfully created: %s", tableName)
		columns, _ := handler.apiService.GetSchema(r.Context(), tableName)
		handler.responseEncoder.EncodeResponse(w, r, http.StatusCreated, columns, format, "columns", nil)

	case http.MethodDelete:
		if err := handler.apiService.DropTable(r.Context(), tableName); err != nil {
			log.Printf("[ERROR] Failed to drop table %s: %v", tableName, err)
			handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err, "Table: "+tableName))
			return
		}
		handler.openAPIGenerator.Invalidate()
		log.Printf("[INFO] Table successfully dropped: %s", tableName)
		w.WriteHeader(http.StatusNoContent)

	default:
		handler.responseEncoder.EncodeError(w, r, errors.ErrMethodNotAllowed)
	}
}

func (handler *APIHandler) handleCRUD(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	requestPath := strings.TrimPrefix(r.URL.Path, handler.basePath)
	pathParts := strings.Split(strings.Trim(requestPath, "/"), "/")

	if len(pathParts) == 0 || pathParts[0] == "" {
		handler.responseEncoder.EncodeError(w, r, errors.ErrNotFound)
		return
	}

	if pathParts[0] == "openapi.json" {
		handler.handleOpenAPI(w, r)
		return
	}

	if pathParts[0] == "docs" {
		handler.handleDocsUI(w, r)
		return
	}

	tableName := pathParts[0]

	// GET /v1/<table>
	if len(pathParts) == 1 && r.Method == http.MethodGet {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		req := domain.ListRequest{
			Limit:  limit,
			Offset: offset,
			Order:  r.URL.Query().Get("order"),
			Where:  r.URL.Query().Get("where"),
		}

		records, err := handler.apiService.List(r.Context(), tableName, req)
		if err != nil {
			if service.IsInvalidColumn(err) {
				handler.responseEncoder.EncodeError(w, r, errors.ErrBadRequest.With(err))
				return
			}
			if service.IsMalformedQuery(err) {
				handler.responseEncoder.EncodeError(w, r, errors.ErrUnprocessableEntity.With(err))
				return
			}
			if err == domain.ErrNotFound {
				handler.responseEncoder.EncodeError(w, r, errors.ErrNotFound.With(err, "Table: "+tableName))
				return
			}
			log.Printf("[ERROR] Failed to list records for table %s: %v", tableName, err)
			handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err, "Table: "+tableName))
			return
		}

		schema, _ := handler.apiService.GetSchema(r.Context(), tableName)
		handler.responseEncoder.EncodeResponse(w, r, http.StatusOK, records, domain.FormatJSON, tableName, schema)
		return
	}

	// /v1/<table>/<id>
	if len(pathParts) == 2 {
		recordID := pathParts[1]

		switch r.Method {

		case http.MethodGet:
			record, err := handler.apiService.GetByID(r.Context(), tableName, recordID)
			if err != nil {
				handler.responseEncoder.EncodeError(w, r, errors.ErrNotFound.With(err, "Record ID: "+recordID))
				return
			}
			schema, _ := handler.apiService.GetSchema(r.Context(), tableName)
			handler.responseEncoder.EncodeResponse(w, r, http.StatusOK, record, domain.FormatJSON, tableName, schema)

		case http.MethodPut, http.MethodPatch:
			var payload map[string]any
			format, decodeErr := DecodeInputPayload(r, &payload)
			if decodeErr != nil {
				handler.responseEncoder.EncodeError(w, r, errors.ErrBadRequest.With(decodeErr))
				return
			}

			record, err := handler.apiService.Update(r.Context(), tableName, recordID, payload)
			if err != nil {
				log.Printf("[ERROR] Failed to update record %s in table %s: %v", recordID, tableName, err)
				handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err, "Record ID: "+recordID))
				return
			}

			schema, _ := handler.apiService.GetSchema(r.Context(), tableName)
			handler.responseEncoder.EncodeResponse(w, r, http.StatusOK, record, format, tableName, schema)

		case http.MethodDelete:
			if err := handler.apiService.Delete(r.Context(), tableName, recordID); err != nil {
				log.Printf("[ERROR] Failed to delete record %s in table %s: %v", recordID, tableName, err)
				handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err, "Record ID: "+recordID))
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			handler.responseEncoder.EncodeError(w, r, errors.ErrMethodNotAllowed)
		}
		return
	}

	// POST /v1/<table>
	if len(pathParts) == 1 && r.Method == http.MethodPost {
		var payload map[string]any
		format, decodeErr := DecodeInputPayload(r, &payload)
		if decodeErr != nil {
			handler.responseEncoder.EncodeError(w, r, errors.ErrBadRequest.With(decodeErr))
			return
		}

		record, err := handler.apiService.Insert(r.Context(), tableName, payload)
		if err != nil {
			log.Printf("[ERROR] Failed to insert record into table %s: %v", tableName, err)
			handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err))
			return
		}

		schema, _ := handler.apiService.GetSchema(r.Context(), tableName)
		handler.responseEncoder.EncodeResponse(w, r, http.StatusCreated, record, format, tableName, schema)
		return
	}

	handler.responseEncoder.EncodeError(w, r, errors.ErrNotFound)
}

func (handler *APIHandler) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		handler.responseEncoder.EncodeError(w, r, errors.ErrMethodNotAllowed)
		return
	}

	data, err := handler.openAPIGenerator.GenerateJSON(r.Context())
	if err != nil {
		log.Printf("[ERROR] Failed to generate OpenAPI spec: %v", err)
		handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(data)
}

func (handler *APIHandler) handleDocsUI(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		handler.responseEncoder.EncodeError(w, r, errors.ErrMethodNotAllowed)
		return
	}

	specURL := strings.TrimRight(handler.basePath, "/") + "/openapi.json"
	docTitle := "Conduit API Reference"
	if handler.serverConfig != nil && handler.serverConfig.OpenAPI.Title != "" {
		docTitle = fmt.Sprintf("%s Reference", handler.serverConfig.OpenAPI.Title)
	}
	html := openapi.DocsHTML(specURL, docTitle)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(html)
}
