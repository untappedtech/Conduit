package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	authPkg "github.com/untappedtech/conduit/internal/auth"
	"github.com/untappedtech/conduit/internal/db/impl"
	"github.com/untappedtech/conduit/internal/domain"
	httpPkg "github.com/untappedtech/conduit/internal/http"
	"github.com/untappedtech/conduit/internal/service"
)

func setupRateLimitedServer(rps float64, burst int, enabled bool) (http.Handler, *httpPkg.RateLimiter) {
	serverConfig := &domain.ServerConfig{}
	serverConfig.Policy.PublicReads = true
	serverConfig.Policy.PublicWrites = true
	serverConfig.RateLimit.Enabled = enabled
	serverConfig.RateLimit.RequestsPerSecond = rps
	serverConfig.RateLimit.Burst = burst

	authChain, _ := authPkg.BuildAuthChain(serverConfig)
	tokenExtractor := authPkg.NewDefaultTokenExtractor()
	responseEncoder := httpPkg.NewResponseEncoder()

	memoryDB := impl.NewMemoryDB()
	apiService := service.NewAPIService(memoryDB, serverConfig)
	apiHandler := httpPkg.NewAPIHandler(apiService, serverConfig, responseEncoder)

	serveMux := http.NewServeMux()
	apiHandler.RegisterRoutes(serveMux)

	rateLimiter := httpPkg.NewRateLimiter(serverConfig.RateLimit, responseEncoder, tokenExtractor)

	handler := authPkg.AuthMiddleware(authChain, tokenExtractor, responseEncoder)(serveMux)
	if enabled {
		handler = rateLimiter.Middleware()(handler)
	}

	return handler, rateLimiter
}

func TestRateLimit_IPLimiting(t *testing.T) {
	handler, rl := setupRateLimitedServer(1, 3, true)
	defer rl.Close()

	clientA := "192.0.2.1:12345"

	// First 3 requests from Client A should succeed (burst = 3)
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
		req.RemoteAddr = clientA
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d from client A expected 200, got %d: %s", i, rec.Code, rec.Body.String())
		}
	}

	// 4th request from Client A should immediately be rejected with 429
	req4 := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
	req4.RemoteAddr = clientA
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests for 4th request from Client A, got %d", rec4.Code)
	}

	var errBody map[string]any
	if err := json.Unmarshal(rec4.Body.Bytes(), &errBody); err != nil {
		t.Fatalf("failed to decode 429 JSON response: %v", err)
	}
	errObj, ok := errBody["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error envelope in 429 response")
	}
	if errObj["status"].(float64) != 429 {
		t.Errorf("expected error status 429, got %v", errObj["status"])
	}
	if errObj["title"].(string) != "Too Many Requests" {
		t.Errorf("expected title 'Too Many Requests', got %v", errObj["title"])
	}

	// Client B (different IP) should NOT be rate limited
	clientB := "198.51.100.2:54321"
	reqB := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
	reqB.RemoteAddr = clientB
	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Fatalf("client B expected 200, got %d: %s", recB.Code, recB.Body.String())
	}
}

func TestRateLimit_TokenLimiting(t *testing.T) {
	handler, rl := setupRateLimitedServer(1, 2, true)
	defer rl.Close()

	tokenAlice := "Bearer alice-token-123"
	tokenBob := "Bearer bob-token-456"

	// Alice sends 2 requests (burst = 2)
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
		req.Header.Set("Authorization", tokenAlice)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Alice request %d expected 200, got %d", i, rec.Code)
		}
	}

	// Alice 3rd request -> 429
	reqAlice3 := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
	reqAlice3.Header.Set("Authorization", tokenAlice)
	recAlice3 := httptest.NewRecorder()
	handler.ServeHTTP(recAlice3, reqAlice3)
	if recAlice3.Code != http.StatusTooManyRequests {
		t.Fatalf("Alice 3rd request expected 429, got %d", recAlice3.Code)
	}

	// Bob sends request from the same IP -> should succeed because token differs
	reqBob := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
	reqBob.Header.Set("Authorization", tokenBob)
	recBob := httptest.NewRecorder()
	handler.ServeHTTP(recBob, reqBob)
	if recBob.Code != http.StatusOK {
		t.Fatalf("Bob request expected 200, got %d", recBob.Code)
	}
}

func TestRateLimit_XForwardedForHeader(t *testing.T) {
	handler, rl := setupRateLimitedServer(1, 2, true)
	defer rl.Close()

	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.195, 10.0.0.1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	req3 := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
	req3.Header.Set("X-Forwarded-For", "203.0.113.195, 10.0.0.1")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for X-Forwarded-For IP, got %d", rec3.Code)
	}
}

func TestRateLimit_DisabledPassesRequests(t *testing.T) {
	handler, rl := setupRateLimitedServer(1, 2, false)
	defer rl.Close()

	// 10 requests should all pass when rate limiting is disabled
	for i := 1; i <= 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/schema", nil)
		req.RemoteAddr = "192.0.2.10:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d expected 200 when rate limiting disabled, got %d", i, rec.Code)
		}
	}
}

func TestRateLimit_StaleSweeper(t *testing.T) {
	config := domain.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10,
		Burst:             10,
	}
	rl := httpPkg.NewRateLimiter(config, nil)
	defer rl.Close()

	rl.SetSweepTimes(10*time.Millisecond, 20*time.Millisecond)

	// Add an entry
	_ = rl.GetLimiter("ip:1.2.3.4")
	if rl.ActiveLimitersCount() != 1 {
		t.Fatalf("expected 1 active limiter, got %d", rl.ActiveLimitersCount())
	}

	// Sleep longer than staleTimeout (20ms)
	time.Sleep(30 * time.Millisecond)

	// Trigger sweep
	rl.Sweep()

	if rl.ActiveLimitersCount() != 0 {
		t.Fatalf("expected 0 active limiters after sweep, got %d", rl.ActiveLimitersCount())
	}
}

func TestHTTP_WhereASTComplexityBouncer(t *testing.T) {
	srv := setupTestHTTPServer()
	createSportsSchema(t, srv)
	insertSportsRecords(t, srv)

	// 1. Valid nesting depth <= 5 (depth 2) -> 200 OK
	reqValid := httptest.NewRequest(http.MethodGet, "/v1/data/sports?where="+url.QueryEscape("((name = 'Golf'))"), nil)
	recValid := httptest.NewRecorder()
	srv.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for depth 2 query, got %d: %s", recValid.Code, recValid.Body.String())
	}

	// 2. Exceeding nesting depth > 5 (depth 6) -> 400 Bad Request
	reqTooDeep := httptest.NewRequest(http.MethodGet, "/v1/data/sports?where="+url.QueryEscape("((((((name = 'Golf'))))))"), nil)
	recTooDeep := httptest.NewRecorder()
	srv.ServeHTTP(recTooDeep, reqTooDeep)
	if recTooDeep.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for depth 6 query, got %d: %s", recTooDeep.Code, recTooDeep.Body.String())
	}

	var errBody map[string]any
	if err := json.Unmarshal(recTooDeep.Body.Bytes(), &errBody); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	errObj := errBody["error"].(map[string]any)
	if errObj["status"].(float64) != 400 {
		t.Errorf("expected status 400, got %v", errObj["status"])
	}
	if errObj["title"].(string) != "Bad Request" {
		t.Errorf("expected title 'Bad Request', got %v", errObj["title"])
	}
}
