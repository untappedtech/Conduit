package httpdecoder_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/untappedtech/conduit/internal/domain"
	httpPkg "github.com/untappedtech/conduit/internal/http"
)

func TestDecoder_XML(t *testing.T) {
	body := bytes.NewBufferString("<Sample><name>alpha</name></Sample>")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/xml")

	var sample Sample
	format, err := httpPkg.DecodeInputPayload(req, &sample)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if sample.Name != "alpha" {
		t.Fatalf("expected name=alpha")
	}
	if format != domain.FormatXML {
		t.Fatalf("expected XML format")
	}
}

func TestDecoder_XMLMap(t *testing.T) {
	xmlData := `<item><name>Eve</name><score>65</score><active>true</active></item>`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(xmlData))
	req.Header.Set("Content-Type", "application/xml")

	var payload map[string]any
	format, err := httpPkg.DecodeInputPayload(req, &payload)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if format != domain.FormatXML {
		t.Fatalf("expected XML format")
	}
	if payload["name"] != "Eve" {
		t.Fatalf("expected name=Eve, got %v", payload["name"])
	}
	if payload["score"] != int64(65) {
		t.Fatalf("expected score=65 (int64), got %v (%T)", payload["score"], payload["score"])
	}
	if payload["active"] != true {
		t.Fatalf("expected active=true, got %v", payload["active"])
	}
}

func TestDecoder_XMLSchema(t *testing.T) {
	xmlData := `<schema>
  <column>
    <name>id</name>
    <type>INTEGER</type>
    <pk>true</pk>
    <autoincrement>true</autoincrement>
  </column>
  <column>
    <name>name</name>
    <type>TEXT</type>
  </column>
</schema>`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(xmlData))
	req.Header.Set("Content-Type", "application/xml")

	var payload struct {
		Columns []domain.ColumnDef `json:"columns" yaml:"columns" xml:"column" toml:"columns"`
	}
	format, err := httpPkg.DecodeInputPayload(req, &payload)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if format != domain.FormatXML {
		t.Fatalf("expected XML format, got %v", format)
	}
	if len(payload.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(payload.Columns))
	}
	if payload.Columns[0].Name != "id" || payload.Columns[1].Name != "name" {
		t.Fatalf("unexpected column names: %v, %v", payload.Columns[0].Name, payload.Columns[1].Name)
	}
}


