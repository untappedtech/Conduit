package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/untappedtech/conduit/internal/config"
)

func TestConfig_OpenAPIDefaults(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")

	// Config without openapi section
	yamlContent := `
server:
  host: "127.0.0.1"
  port: 8080
database:
  driver: "sqlite"
  dsn: ":memory:"
`
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.OpenAPI.Title != "Conduit API" {
		t.Fatalf("expected default title 'Conduit API', got %q", cfg.OpenAPI.Title)
	}
	if cfg.OpenAPI.Version != "1.0.0" {
		t.Fatalf("expected default version '1.0.0', got %q", cfg.OpenAPI.Version)
	}
	if !strings.Contains(cfg.OpenAPI.Description, "Multi-Format") {
		t.Fatalf("expected default description, got %q", cfg.OpenAPI.Description)
	}
}

func TestConfig_OpenAPICustomYAML(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `
server:
  host: "0.0.0.0"
  port: 9000
database:
  driver: "sqlite"
  dsn: ":memory:"
openapi:
  title: "Inventory API"
  version: "2.5.0"
  description: "Warehouse inventory management API"
  terms_of_service: "https://warehouse.example.com/terms"
  contact:
    name: "DevOps"
    email: "devops@warehouse.example.com"
    url: "https://warehouse.example.com/support"
  license:
    name: "Apache-2.0"
    url: "https://www.apache.org/licenses/LICENSE-2.0"
  servers:
    - url: "https://warehouse.example.com/v1"
      description: "Prod Gateway"
    - url: "http://localhost:9000/v1"
      description: "Local Dev"
`
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.OpenAPI.Title != "Inventory API" {
		t.Fatalf("expected 'Inventory API', got %q", cfg.OpenAPI.Title)
	}
	if cfg.OpenAPI.Version != "2.5.0" {
		t.Fatalf("expected '2.5.0', got %q", cfg.OpenAPI.Version)
	}
	if cfg.OpenAPI.Description != "Warehouse inventory management API" {
		t.Fatalf("expected custom description, got %q", cfg.OpenAPI.Description)
	}
	if cfg.OpenAPI.TermsOfService != "https://warehouse.example.com/terms" {
		t.Fatalf("expected custom terms, got %q", cfg.OpenAPI.TermsOfService)
	}
	if cfg.OpenAPI.Contact == nil || cfg.OpenAPI.Contact.Name != "DevOps" || cfg.OpenAPI.Contact.Email != "devops@warehouse.example.com" {
		t.Fatalf("expected custom contact, got %+v", cfg.OpenAPI.Contact)
	}
	if cfg.OpenAPI.License == nil || cfg.OpenAPI.License.Name != "Apache-2.0" {
		t.Fatalf("expected custom license, got %+v", cfg.OpenAPI.License)
	}
	if len(cfg.OpenAPI.Servers) != 2 || cfg.OpenAPI.Servers[0].URL != "https://warehouse.example.com/v1" {
		t.Fatalf("expected 2 servers, got %+v", cfg.OpenAPI.Servers)
	}
}

func TestConfig_OpenAPICustomJSON(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.json")

	jsonContent := `{
  "server": { "host": "127.0.0.1", "port": 8080 },
  "database": { "driver": "sqlite", "dsn": ":memory:" },
  "openapi": {
    "title": "JSON API",
    "version": "1.2.3",
    "termsOfService": "https://json.example.com/terms",
    "contact": { "name": "Team", "email": "team@json.example.com" },
    "license": { "name": "MIT" }
  }
}`
	if err := os.WriteFile(cfgPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.OpenAPI.Title != "JSON API" {
		t.Fatalf("expected 'JSON API', got %q", cfg.OpenAPI.Title)
	}
	if cfg.OpenAPI.Version != "1.2.3" {
		t.Fatalf("expected '1.2.3', got %q", cfg.OpenAPI.Version)
	}
	if cfg.OpenAPI.TermsOfService != "https://json.example.com/terms" {
		t.Fatalf("expected camelCase termsOfService match, got %q", cfg.OpenAPI.TermsOfService)
	}
	if cfg.OpenAPI.Contact == nil || cfg.OpenAPI.Contact.Name != "Team" {
		t.Fatalf("expected contact match, got %+v", cfg.OpenAPI.Contact)
	}
	if cfg.OpenAPI.License == nil || cfg.OpenAPI.License.Name != "MIT" {
		t.Fatalf("expected license match, got %+v", cfg.OpenAPI.License)
	}
}

func TestConfig_GenerateDefaultConfigOpenAPI(t *testing.T) {
	tempDir := t.TempDir()
	for _, format := range []string{"yaml", "json", "toml"} {
		outPath := filepath.Join(tempDir, "config."+format)
		if err := config.GenerateDefaultConfig(format, outPath); err != nil {
			t.Fatalf("failed to generate %s config: %v", format, err)
		}
		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read %s config: %v", format, err)
		}
		content := string(data)
		if !strings.Contains(content, "Conduit API") {
			t.Fatalf("expected %s config to contain 'Conduit API', got:\n%s", format, content)
		}
	}
}
