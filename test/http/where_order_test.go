package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func queryURL(base, where, order string) string {
	u := base
	sep := "?"
	if where != "" {
		u += sep + "where=" + url.QueryEscape(where)
		sep = "&"
	}
	if order != "" {
		u += sep + "order=" + url.QueryEscape(order)
	}
	return u
}

func insertSportsForQuery(t *testing.T, srv http.Handler) {
	createSportsSchema(t, srv)

	sports := []struct {
		name    string
		players int
	}{
		{"Golf", 1},
		{"Tennis", 2},
		{"Basketball", 5},
		{"Baseball", 9},
		{"Soccer", 11},
	}

	for _, s := range sports {
		rec := httptest.NewRecorder()
		body := `{"name":"` + s.name + `","players":` + jsonNum(s.players) + `}`
		req := httptest.NewRequest(http.MethodPost, "/v1/sports", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("failed to insert sport %s: code %d", s.name, rec.Code)
		}
	}
}

func jsonNum(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func TestHTTP_Order_AscendingAndDescending(t *testing.T) {
	srv := setupTestHTTPServer()
	insertSportsForQuery(t, srv)

	// Test ASC
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "", "name:asc"), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Sports []struct {
				Name string `json:"name"`
			} `json:"sports"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if len(payload.Sports) != 5 {
			t.Fatalf("expected 5 sports, got %d", len(payload.Sports))
		}
		if payload.Sports[0].Name != "Baseball" {
			t.Errorf("expected first sport to be Baseball, got %s", payload.Sports[0].Name)
		}
		if payload.Sports[4].Name != "Tennis" {
			t.Errorf("expected last sport to be Tennis, got %s", payload.Sports[4].Name)
		}
	}

	// Test DESC
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "", "name:desc"), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var payload struct {
			Sports []struct {
				Name string `json:"name"`
			} `json:"sports"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
		if payload.Sports[0].Name != "Tennis" {
			t.Errorf("expected first sport to be Tennis, got %s", payload.Sports[0].Name)
		}
		if payload.Sports[4].Name != "Baseball" {
			t.Errorf("expected last sport to be Baseball, got %s", payload.Sports[4].Name)
		}
	}

	// Test Numeric DESC
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "", "players:desc"), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var payload struct {
			Sports []struct {
				Name    string `json:"name"`
				Players int    `json:"players"`
			} `json:"sports"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
		if payload.Sports[0].Name != "Soccer" || payload.Sports[0].Players != 11 {
			t.Errorf("expected first sport to be Soccer (11), got %s (%d)", payload.Sports[0].Name, payload.Sports[0].Players)
		}
		if payload.Sports[4].Name != "Golf" || payload.Sports[4].Players != 1 {
			t.Errorf("expected last sport to be Golf (1), got %s (%d)", payload.Sports[4].Name, payload.Sports[4].Players)
		}
	}
}

func TestHTTP_Where_Filters(t *testing.T) {
	srv := setupTestHTTPServer()
	insertSportsForQuery(t, srv)

	// Filter: players > 5
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "players > 5", ""), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Sports []struct {
				Name    string `json:"name"`
				Players int    `json:"players"`
			} `json:"sports"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
		if len(payload.Sports) != 2 {
			t.Fatalf("expected 2 sports with players > 5, got %d", len(payload.Sports))
		}
		for _, s := range payload.Sports {
			if s.Players <= 5 {
				t.Errorf("unexpected sport %s with %d players", s.Name, s.Players)
			}
		}
	}

	// Filter: name = 'Golf'
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "name = 'Golf'", ""), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Sports []struct {
				Name string `json:"name"`
			} `json:"sports"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
		if len(payload.Sports) != 1 || payload.Sports[0].Name != "Golf" {
			t.Fatalf("expected only Golf, got %#v", payload.Sports)
		}
	}

	// Filter: name LIKE '%ball%'
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "name LIKE '%ball%'", ""), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Sports []struct {
				Name string `json:"name"`
			} `json:"sports"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
		if len(payload.Sports) != 2 {
			t.Fatalf("expected 2 ball sports (Basketball, Baseball), got %d", len(payload.Sports))
		}
	}

	// Filter: name IN ('Tennis', 'Soccer')
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "name IN ('Tennis', 'Soccer')", ""), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Sports []struct {
				Name string `json:"name"`
			} `json:"sports"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
		if len(payload.Sports) != 2 {
			t.Fatalf("expected 2 sports, got %d", len(payload.Sports))
		}
	}

	// Combined: where AND order
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "players >= 2", "players:desc"), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Sports []struct {
				Name    string `json:"name"`
				Players int    `json:"players"`
			} `json:"sports"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
		if len(payload.Sports) != 4 {
			t.Fatalf("expected 4 sports, got %d", len(payload.Sports))
		}
		if payload.Sports[0].Name != "Soccer" {
			t.Errorf("expected first sport to be Soccer, got %s", payload.Sports[0].Name)
		}
		if payload.Sports[3].Name != "Tennis" {
			t.Errorf("expected last sport to be Tennis, got %s", payload.Sports[3].Name)
		}
	}
}

func TestHTTP_OrderAndWhere_ErrorValidation(t *testing.T) {
	srv := setupTestHTTPServer()
	insertSportsForQuery(t, srv)

	// Invalid column in order -> 400 Bad Request
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "", "non_existent_col:asc"), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid order column, got %d", rec.Code)
		}
	}

	// Invalid direction in order -> 422 Unprocessable Entity
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "", "name:backwards"), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422 for invalid order direction, got %d", rec.Code)
		}
	}

	// Invalid column in where -> 400 Bad Request
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "non_existent > 10", ""), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid where column, got %d", rec.Code)
		}
	}

	// Malformed syntax in where -> 422 Unprocessable Entity
	{
		req := httptest.NewRequest(http.MethodGet, queryURL("/v1/sports", "players >", ""), nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422 for malformed where syntax, got %d", rec.Code)
		}
	}
}
