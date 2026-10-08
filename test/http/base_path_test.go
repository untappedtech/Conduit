package http_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authPkg "github.com/untappedtech/conduit/internal/auth"
	"github.com/untappedtech/conduit/internal/config"
	"github.com/untappedtech/conduit/internal/db/impl"
	"github.com/untappedtech/conduit/internal/domain"
	httpPkg "github.com/untappedtech/conduit/internal/http"
	"github.com/untappedtech/conduit/internal/service"
)

func setupTestServerWithBasePath(basePath string) (http.Handler, *domain.ServerConfig) {
	serverConfig := &domain.ServerConfig{}
	serverConfig.Server.DefaultLimit = 50
	serverConfig.Server.BasePath = basePath
	serverConfig.Policy.PublicReads = true
	serverConfig.Policy.PublicWrites = true
	serverConfig.Policy.PublicMutation = true

	authChain, _ := authPkg.BuildAuthChain(serverConfig)
	tokenExtractor := authPkg.NewDefaultTokenExtractor()
	responseEncoder := httpPkg.NewResponseEncoder()

	memoryDB := impl.NewMemoryDB()
	apiService := service.NewAPIService(memoryDB, serverConfig)
	server := httpPkg.NewServer(apiService, serverConfig, responseEncoder)

	handler := authPkg.AuthMiddleware(authChain, tokenExtractor, responseEncoder)(server)
	return handler, serverConfig
}

func TestConfigurableBasePath_Default(t *testing.T) {
	serverConfig := &domain.ServerConfig{}
	handler := httpPkg.NewAPIHandler(nil, serverConfig, httpPkg.NewResponseEncoder())
	if handler.BasePath() != "/v1/" {
		t.Fatalf("expected default /v1/, got %q", handler.BasePath())
	}
}

func TestConfigurableBasePath_Custom(t *testing.T) {
	customPath := "/api/v2/"
	srv, _ := setupTestServerWithBasePath(customPath)

	// 1. Create schema on /api/v2/schema/items
	createSchemaBody := `{
		"columns": [
			{"name": "id", "type": "INTEGER", "pk": true, "autoincrement": true},
			{"name": "title", "type": "TEXT"}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/schema/items", strings.NewReader(createSchemaBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("create schema failed with code %d: %s", rec.Code, rec.Body.String())
	}

	// 2. List tables on /api/v2/schema
	req = httptest.NewRequest(http.MethodGet, "/api/v2/schema", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list tables failed with code %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "items") {
		t.Fatalf("expected response to contain 'items', got %s", rec.Body.String())
	}

	// 3. Insert record on /api/v2/items
	req = httptest.NewRequest(http.MethodPost, "/api/v2/items", strings.NewReader(`{"title": "Item 1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("insert failed with code %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Query record on /api/v2/items
	req = httptest.NewRequest(http.MethodGet, "/api/v2/items", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list items failed with code %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Item 1") {
		t.Fatalf("expected 'Item 1' in response, got %s", rec.Body.String())
	}

	// 5. Query /v1/items should 404 because BasePath is /api/v2/
	req = httptest.NewRequest(http.MethodGet, "/v1/items", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for old path /v1/items, got %d", rec.Code)
	}
}

func TestConfigurableBasePath_Normalization(t *testing.T) {
	// Without leading and trailing slashes: "custom/api" -> "/custom/api/"
	srv, _ := setupTestServerWithBasePath("custom/api")

	req := httptest.NewRequest(http.MethodGet, "/custom/api/schema", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for normalized base path, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConfigParsing_BasePathHandling(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Test parsing YAML with explicit base_path
	yamlPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := `
server:
    host: "127.0.0.1"
    port: 9000
    base_path: "/custom/api/"
database:
    driver: "sqlite"
    dsn: ":memory:"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	cfg, err := config.Load(yamlPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Server.BasePath != "/custom/api/" {
		t.Fatalf("expected /custom/api/, got %q", cfg.Server.BasePath)
	}

	// 2. Test parsing YAML without base_path (defaults to /v1/)
	noBasePathYaml := filepath.Join(tempDir, "nobase.yaml")
	noBaseContent := `
server:
    host: "127.0.0.1"
    port: 9000
database:
    driver: "sqlite"
    dsn: ":memory:"
`
	if err := os.WriteFile(noBasePathYaml, []byte(noBaseContent), 0644); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	cfgNoBase, err := config.Load(noBasePathYaml)
	if err != nil {
		t.Fatalf("failed to load config without base_path: %v", err)
	}
	if cfgNoBase.Server.BasePath != "/v1/" {
		t.Fatalf("expected default /v1/, got %q", cfgNoBase.Server.BasePath)
	}

	// 3. Test GenerateDefaultConfig includes base_path in generated file
	formats := []string{"yaml", "json", "toml"}
	for _, fmtType := range formats {
		outPath := filepath.Join(tempDir, "gen."+fmtType)
		if err := config.GenerateDefaultConfig(fmtType, outPath); err != nil {
			t.Fatalf("failed to generate %s config: %v", fmtType, err)
		}
		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read generated %s config: %v", fmtType, err)
		}
		if !strings.Contains(string(data), "/v1/") {
			t.Fatalf("expected generated %s config to contain '/v1/', got:\n%s", fmtType, string(data))
		}
	}
}
