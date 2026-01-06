package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
)

func TestLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	app := kese.New()
	app.Use(Logger())

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	logOutput := buf.String()

	// Check that log contains method, path, and status
	if !strings.Contains(logOutput, "GET") {
		t.Error("Log should contain HTTP method")
	}
	if !strings.Contains(logOutput, "/test") {
		t.Error("Log should contain path")
	}
	if !strings.Contains(logOutput, "200") {
		t.Error("Log should contain status code")
	}
}

func TestRecovery(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	app := kese.New()
	app.Use(Recovery())

	app.GET("/panic", func(c *context.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// Should not panic
	app.ServeHTTP(w, req)

	// Should return 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Should contain error response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response["error"] != "Internal Server Error" {
		t.Error("Response should contain error message")
	}

	// Log should contain panic info
	logOutput := buf.String()
	if !strings.Contains(logOutput, "PANIC") {
		t.Error("Log should contain PANIC message")
	}
	if !strings.Contains(logOutput, "test panic") {
		t.Error("Log should contain panic message")
	}
}

func TestRecoveryDoesNotAffectNormalRequests(t *testing.T) {
	app := kese.New()
	app.Use(Recovery())

	app.GET("/normal", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/normal", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != "OK" {
		t.Errorf("Expected body 'OK', got %q", w.Body.String())
	}
}

func TestCORS(t *testing.T) {
	app := kese.New()
	app.Use(CORS())

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Access-Control-Allow-Origin header should be set to *")
	}

	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "GET") || !strings.Contains(allowMethods, "POST") {
		t.Errorf("Access-Control-Allow-Methods should contain GET and POST, got %s", allowMethods)
	}

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(allowHeaders, "Content-Type") {
		t.Errorf("Access-Control-Allow-Headers should contain Content-Type, got %s", allowHeaders)
	}
}

func TestCORSPreflight(t *testing.T) {
	app := kese.New()
	app.Use(CORS())

	// Register OPTIONS route explicitly
	app.OPTIONS("/test", func(c *context.Context) error {
		return c.String(200, "Should not reach here")
	})

	// Send OPTIONS request (preflight)
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Should return 204 No Content (from CORS middleware)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", w.Code)
	}

	// Should have CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS headers should be set for preflight")
	}
}

func TestCORSWithConfig(t *testing.T) {
	app := kese.New()
	app.Use(CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Authorization"},
	}))

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Check custom origin
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected origin https://example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Check custom methods
	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "GET") || !strings.Contains(allowMethods, "POST") {
		t.Errorf("Expected GET, POST in methods, got %s", allowMethods)
	}

	// Check custom headers
	if w.Header().Get("Access-Control-Allow-Headers") != "Authorization" {
		t.Errorf("Expected Authorization in headers, got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestRequestID(t *testing.T) {
	app := kese.New()
	app.Use(RequestID())

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Check that request ID header is set
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header should be set")
	}

	// Request ID should contain timestamp and counter
	if !strings.Contains(requestID, "-") {
		t.Error("Request ID should contain timestamp-counter format")
	}
}

func TestRequestIDIncrement(t *testing.T) {
	app := kese.New()
	app.Use(RequestID())

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	// Make multiple requests
	var requestIDs []string
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		requestIDs = append(requestIDs, requestID)
	}

	// All request IDs should be different
	seen := make(map[string]bool)
	for _, id := range requestIDs {
		if seen[id] {
			t.Errorf("Request ID %s was duplicated", id)
		}
		seen[id] = true
	}
}

func TestRequestIDConcurrent(t *testing.T) {
	app := kese.New()
	app.Use(RequestID())

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	// Test concurrent requests to ensure no race conditions
	var wg sync.WaitGroup
	var mu sync.Mutex
	requestIDs := make(map[string]bool)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			app.ServeHTTP(w, req)

			requestID := w.Header().Get("X-Request-ID")

			mu.Lock()
			if requestIDs[requestID] {
				t.Errorf("Duplicate request ID in concurrent test: %s", requestID)
			}
			requestIDs[requestID] = true
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Should have 100 unique request IDs
	if len(requestIDs) != 100 {
		t.Errorf("Expected 100 unique request IDs, got %d", len(requestIDs))
	}
}

func TestMiddlewareChaining(t *testing.T) {
	// Capture log output to avoid nil pointer issues
	var buf bytes.Buffer
	oldOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOutput)

	app := kese.New()

	// Chain multiple middleware
	app.Use(Logger(), Recovery(), CORS(), RequestID())

	app.GET("/test", func(c *context.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// All middleware should have executed
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// CORS headers should be present
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS middleware should have set headers")
	}

	// Request ID should be present
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("RequestID middleware should have set header")
	}
}
