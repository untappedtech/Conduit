package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTP_OpenAPI_Endpoints(t *testing.T) {
	srv := setupTestHTTPServer()

	// 1. GET /v1/openapi.json
	req1 := httptest.NewRequest(http.MethodGet, "/v1/openapi.json", nil)
	rec1 := httptest.NewRecorder()
	srv.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 from /v1/openapi.json, got %d, body: %s", rec1.Code, rec1.Body.String())
	}
	if ct := rec1.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected application/json Content-Type, got %s", ct)
	}

	var spec1 map[string]any
	if err := json.Unmarshal(rec1.Body.Bytes(), &spec1); err != nil {
		t.Fatalf("failed to parse /v1/openapi.json body: %v", err)
	}
	if spec1["openapi"] != "3.0.3" {
		t.Fatalf("expected openapi 3.0.3, got %v", spec1["openapi"])
	}

	// 2. GET /openapi.json (root alias)
	req2 := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 from /openapi.json, got %d", rec2.Code)
	}

	// 3. GET /v1/docs
	req3 := httptest.NewRequest(http.MethodGet, "/v1/docs", nil)
	rec3 := httptest.NewRecorder()
	srv.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 from /v1/docs, got %d", rec3.Code)
	}
	if ct := rec3.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html Content-Type, got %s", ct)
	}
	bodyDocs := rec3.Body.String()
	if !strings.Contains(bodyDocs, "@scalar/api-reference") {
		t.Fatalf("expected scalar script in docs html, got: %s", bodyDocs)
	}

	// 4. GET /docs (root alias)
	req4 := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec4 := httptest.NewRecorder()
	srv.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("expected 200 from /docs, got %d", rec4.Code)
	}
}

func TestHTTP_OpenAPI_DynamicSchemaMutation(t *testing.T) {
	srv := setupTestHTTPServer()

	// Initial spec check: table 'movies' should not exist
	reqInit := httptest.NewRequest(http.MethodGet, "/v1/openapi.json", nil)
	recInit := httptest.NewRecorder()
	srv.ServeHTTP(recInit, reqInit)
	var specInit map[string]any
	_ = json.Unmarshal(recInit.Body.Bytes(), &specInit)
	pathsInit := specInit["paths"].(map[string]any)
	if _, exists := pathsInit["/movies"]; exists {
		t.Fatalf("table /movies should not exist in initial spec")
	}

	// Create table 'movies' via /v1/schema/movies
	createBody := `{
		"columns": [
			{"name": "id", "type": "INTEGER", "pk": true, "autoincrement": true},
			{"name": "title", "type": "VARCHAR(255)", "nullable": false},
			{"name": "year", "type": "INTEGER", "nullable": true}
		]
	}`
	reqCreate := httptest.NewRequest(http.MethodPost, "/v1/schema/movies", strings.NewReader(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	recCreate := httptest.NewRecorder()
	srv.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated && recCreate.Code != http.StatusOK {
		t.Fatalf("failed to create movies table: %d, body: %s", recCreate.Code, recCreate.Body.String())
	}

	// Spec should now reflect 'movies' dynamically
	reqAfterCreate := httptest.NewRequest(http.MethodGet, "/v1/openapi.json", nil)
	recAfterCreate := httptest.NewRecorder()
	srv.ServeHTTP(recAfterCreate, reqAfterCreate)
	var specAfterCreate map[string]any
	_ = json.Unmarshal(recAfterCreate.Body.Bytes(), &specAfterCreate)
	pathsAfterCreate := specAfterCreate["paths"].(map[string]any)
	if _, exists := pathsAfterCreate["/movies"]; !exists {
		t.Fatalf("expected /movies to exist in updated openapi spec")
	}
	if _, exists := pathsAfterCreate["/movies/{id}"]; !exists {
		t.Fatalf("expected /movies/{id} to exist in updated openapi spec")
	}

	// Verify 'where' parameter is present on GET /movies
	moviesPath := pathsAfterCreate["/movies"].(map[string]any)
	getOp := moviesPath["get"].(map[string]any)
	params := getOp["parameters"].([]any)
	hasWhere := false
	for _, p := range params {
		pm := p.(map[string]any)
		if pm["name"] == "where" {
			hasWhere = true
			break
		}
	}
	if !hasWhere {
		t.Fatalf("expected 'where' query parameter on GET /movies")
	}

	// Verify components.schemas contains 'movies' and 'moviesInput'
	components := specAfterCreate["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	if _, exists := schemas["movies"]; !exists {
		t.Fatalf("expected movies in components.schemas")
	}
	if _, exists := schemas["moviesInput"]; !exists {
		t.Fatalf("expected moviesInput in components.schemas")
	}

	// Drop table 'movies'
	reqDrop := httptest.NewRequest(http.MethodDelete, "/v1/schema/movies", nil)
	recDrop := httptest.NewRecorder()
	srv.ServeHTTP(recDrop, reqDrop)
	if recDrop.Code != http.StatusNoContent {
		t.Fatalf("failed to drop movies table: %d", recDrop.Code)
	}

	// Spec should no longer reflect 'movies'
	reqAfterDrop := httptest.NewRequest(http.MethodGet, "/v1/openapi.json", nil)
	recAfterDrop := httptest.NewRecorder()
	srv.ServeHTTP(recAfterDrop, reqAfterDrop)
	var specAfterDrop map[string]any
	_ = json.Unmarshal(recAfterDrop.Body.Bytes(), &specAfterDrop)
	pathsAfterDrop := specAfterDrop["paths"].(map[string]any)
	if _, exists := pathsAfterDrop["/movies"]; exists {
		t.Fatalf("expected /movies to be removed from openapi spec after drop")
	}
}

func TestHTTP_OpenAPI_MethodNotAllowed(t *testing.T) {
	srv := setupTestHTTPServer()

	reqPost := httptest.NewRequest(http.MethodPost, "/v1/openapi.json", strings.NewReader("{}"))
	recPost := httptest.NewRecorder()
	srv.ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed for POST /v1/openapi.json, got %d", recPost.Code)
	}

	reqDocsPost := httptest.NewRequest(http.MethodPost, "/v1/docs", strings.NewReader("{}"))
	recDocsPost := httptest.NewRecorder()
	srv.ServeHTTP(recDocsPost, reqDocsPost)
	if recDocsPost.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed for POST /v1/docs, got %d", recDocsPost.Code)
	}
}
