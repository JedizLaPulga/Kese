package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
)

func TestRateLimit(t *testing.T) {
	app := kese.New()
	// Limit to 2 requests per second
	app.Use(RateLimit(2, time.Second))

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	// 1st request - OK
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("Req 1: Expected 200, got %d", w1.Code)
	}

	// 2nd request - OK
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("Req 2: Expected 200, got %d", w2.Code)
	}

	// 3rd request - Blocked
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, req3)
	if w3.Code != 429 { // Too Many Requests
		t.Errorf("Req 3: Expected 429, got %d", w3.Code)
	}

	// Check Headers
	if w1.Header().Get("X-RateLimit-Limit") != "2" {
		t.Errorf("Expected X-RateLimit-Limit=2")
	}
	remaining := w1.Header().Get("X-RateLimit-Remaining")
	if remaining != "1" { // 2 - 1 = 1 remaining after first
		t.Errorf("Expected X-RateLimit-Remaining=1, got %s", remaining)
	}
}

func TestRateLimitSecurity(t *testing.T) {
	// Verify that X-Forwarded-For is ignored by default
	app := kese.New()
	app.Use(RateLimit(1, time.Minute))

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	// Request 1 from "IP-A"
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Errorf("Req 1 should pass")
	}

	// Request 2 from "IP-A" (spoofing X-Forwarded-For)
	// If it respected X-Forwarded-For, this might pass if the logic was flawed (e.g. using XFF as key)
	// But since we use RemoteAddr, it should be blocked because RemoteAddr is still 1.2.3.4
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "1.2.3.4:5678"              // Same IP
	req2.Header.Set("X-Forwarded-For", "5.6.7.8") // Spoofed IP
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)

	if w2.Code != 429 {
		t.Errorf("Req 2 should be blocked because RemoteAddr matches, even if XFF allows it")
	}

	// Request 3 from "IP-B"
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "9.9.9.9:1234" // Different IP
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, req3)
	if w3.Code != 200 {
		t.Errorf("Req 3 should pass for different IP")
	}
}

func TestRateLimitCustomKey(t *testing.T) {
	// Verify we can override KeyFunc to use X-Forwarded-For if we really want to (behind trusted proxy)
	app := kese.New()

	config := DefaultRateLimitConfig(1, time.Minute)
	config.KeyFunc = func(c *context.Context) string {
		return c.Header("X-Forwarded-For")
	}

	app.Use(RateLimitWithConfig(config))

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	// Req 1 - IP A
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "1.1.1.1")
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Error("Req 1 failed")
	}

	// Req 2 - IP A (Blocked)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "1.1.1.1")
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)
	if w2.Code != 429 {
		t.Error("Req 2 should be blocked")
	}

	// Req 3 - IP B (Allowed)
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Forwarded-For", "2.2.2.2")
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, req3)
	if w3.Code != 200 {
		t.Error("Req 3 failed")
	}
}
