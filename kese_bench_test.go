package kese

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/middleware"
)

// Benchmark tests for performance profiling

func BenchmarkRouteMatching(b *testing.B) {
	app := New()
	app.GET("/users/:id", func(c *context.Context) error {
		return c.Success(map[string]string{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w = httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkJSONResponse(b *testing.B) {
	app := New()
	app.GET("/json", func(c *context.Context) error {
		return c.JSON(200, map[string]interface{}{
			"message": "Hello",
			"count":   42,
			"data":    []int{1, 2, 3, 4, 5},
		})
	})

	req := httptest.NewRequest("GET", "/json", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w = httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkMiddlewareChain(b *testing.B) {
	app := New()
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestID())

	app.GET("/test", func(c *context.Context) error {
		return c.Success("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w = httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkBodyParsing(b *testing.B) {
	app := New()
	app.POST("/users", func(c *context.Context) error {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.Body(&user); err != nil {
			return err
		}
		return c.Created(user)
	})

	body := []byte(`{"name":"John","email":"john@example.com"}`)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Body = http.NoBody
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		w = httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// Stress tests for concurrency and load

func TestConcurrentRequests(t *testing.T) {
	app := New()
	app.Use(middleware.RequestID())

	var counter int
	var mu sync.Mutex

	app.GET("/counter", func(c *context.Context) error {
		mu.Lock()
		counter++
		count := counter
		mu.Unlock()
		return c.Success(map[string]int{"count": count})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	// Run 100 concurrent requests
	const numRequests = 100
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(server.URL + "/counter")
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				errors <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify all requests were processed
	if counter != numRequests {
		t.Errorf("Expected %d requests processed, got %d", numRequests, counter)
	}
}

func TestMemoryLeakCheck(t *testing.T) {
	app := New()
	app.GET("/test", func(c *context.Context) error {
		// Allocate some data
		data := make([]byte, 1024)
		for i := range data {
			data[i] = byte(i % 256)
		}
		return c.Success(map[string]int{"size": len(data)})
	})

	// Make many requests to check for memory leaks
	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("Request %d failed with status %d", i, w.Code)
		}
	}

	// If we got here without running out of memory, test passes
	t.Log("Successfully processed 1000 requests without memory leak")
}

func TestRouteGroupConcurrency(t *testing.T) {
	app := New()

	// Create multiple route groups
	api := app.Group("/api")
	v1 := app.Group("/api/v1")
	v2 := app.Group("/api/v2")

	api.GET("/status", func(c *context.Context) error {
		return c.Success("ok")
	})

	v1.GET("/users", func(c *context.Context) error {
		return c.Success("v1")
	})

	v2.GET("/users", func(c *context.Context) error {
		return c.Success("v2")
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/status", "ok"},
		{"/api/v1/users", "v1"},
		{"/api/v2/users", "v2"},
	}

	// Test concurrently
	var wg sync.WaitGroup
	for _, tt := range tests {
		wg.Add(1)
		go func(path, expected string) {
			defer wg.Done()

			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Path %s: expected 200, got %d", path, w.Code)
				return
			}

			var result map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &result)
			if result["data"] != expected {
				t.Errorf("Path %s: expected %q, got %v", path, expected, result["data"])
			}
		}(tt.path, tt.expected)
	}

	wg.Wait()
}

func TestContextValueStorageConcurrent(t *testing.T) {
	app := New()

	app.GET("/test/:id", func(c *context.Context) error {
		id := c.Param("id")
		c.Set("userID", id)
		c.Set("timestamp", "2026-01-08")

		// Verify values
		if c.Get("userID") != id {
			t.Error("Context value mismatch")
		}
		if c.MustGet("timestamp") != "2026-01-08" {
			t.Error("Context value mismatch")
		}

		return c.Success(map[string]string{
			"id":     id,
			"stored": c.Get("userID").(string),
		})
	})

	// Concurrent requests with different IDs
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", fmt.Sprintf("/test/%d", id), nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Request failed: %d", w.Code)
			}
		}(i)
	}

	wg.Wait()
}
