package http

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/errors"
	"github.com/untappedtech/conduit/internal/service"
)

type APIHandler struct {
	apiService      *service.APIService
	serverConfig    *domain.ServerConfig
	responseEncoder *ResponseEncoder
}

func NewAPIHandler(apiService *service.APIService, serverConfig *domain.ServerConfig, responseEncoder *ResponseEncoder) *APIHandler {
	return &APIHandler{
		apiService:      apiService,
		serverConfig:    serverConfig,
		responseEncoder: responseEncoder,
	}
}

func (handler *APIHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("/v1/schema", handler.handleSchema)
	serveMux.HandleFunc("/v1/schema/", handler.handleSchema)
	serveMux.HandleFunc("/v1/", handler.handleCRUD)
}

func (handler *APIHandler) handleSchema(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	requestPath := strings.TrimPrefix(r.URL.Path, "/v1/schema")
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
		handler.responseEncoder.EncodeError(w, r, errors.ErrBadRequest)
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

		log.Printf("[INFO] Table successfully created: %s", tableName)
		columns, _ := handler.apiService.GetSchema(r.Context(), tableName)
		handler.responseEncoder.EncodeResponse(w, r, http.StatusCreated, columns, format, "columns", nil)

	case http.MethodDelete:
		if err := handler.apiService.DropTable(r.Context(), tableName); err != nil {
			log.Printf("[ERROR] Failed to drop table %s: %v", tableName, err)
			handler.responseEncoder.EncodeError(w, r, errors.ErrInternalServerError.With(err, "Table: "+tableName))
			return
		}
		log.Printf("[INFO] Table successfully dropped: %s", tableName)
		w.WriteHeader(http.StatusNoContent)

	default:
		handler.responseEncoder.EncodeError(w, r, errors.ErrMethodNotAllowed)
	}
}

func (handler *APIHandler) handleCRUD(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	requestPath := strings.TrimPrefix(r.URL.Path, "/v1/")
	pathParts := strings.Split(strings.Trim(requestPath, "/"), "/")

	if len(pathParts) == 0 || pathParts[0] == "" {
		handler.responseEncoder.EncodeError(w, r, errors.ErrNotFound)
		return
	}

	tableName := pathParts[0]

	// GET /v1/<table>
	if len(pathParts) == 1 && r.Method == http.MethodGet {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		records, err := handler.apiService.List(r.Context(), tableName, limit, offset)
		if err != nil {
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
