package kese

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JedizLaPulga/kese/context"
)

func TestNew(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatal("New() returned nil")
	}
	if app.router == nil {
		t.Fatal("App router not initialized")
	}
	if app.middleware == nil {
		t.Fatal("App middleware not initialized")
	}
}

func TestRouteRegistration(t *testing.T) {
	app := New()

	// Register routes
	app.GET("/get", func(c *context.Context) error {
		return c.String(200, "GET")
	})
	app.POST("/post", func(c *context.Context) error {
		return c.String(200, "POST")
	})
	app.PUT("/put", func(c *context.Context) error {
		return c.String(200, "PUT")
	})
	app.DELETE("/delete", func(c *context.Context) error {
		return c.String(200, "DELETE")
	})
	app.PATCH("/patch", func(c *context.Context) error {
		return c.String(200, "PATCH")
	})
	app.OPTIONS("/options", func(c *context.Context) error {
		return c.String(200, "OPTIONS")
	})
	app.HEAD("/head", func(c *context.Context) error {
		return c.NoContent()
	})

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/get", "GET"},
		{"POST", "/post", "POST"},
		{"PUT", "/put", "PUT"},
		{"DELETE", "/delete", "DELETE"},
		{"PATCH", "/patch", "PATCH"},
		{"OPTIONS", "/options", "OPTIONS"},
		{"HEAD", "/head", ""},
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.path, nil)
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		if test.method == "HEAD" {
			if w.Code != http.StatusNoContent {
				t.Errorf("%s %s: expected status 204, got %d", test.method, test.path, w.Code)
			}
		} else {
			if w.Code != http.StatusOK {
				t.Errorf("%s %s: expected status 200, got %d", test.method, test.path, w.Code)
			}
			if strings.TrimSpace(w.Body.String()) != test.expected {
				t.Errorf("%s %s: expected body %q, got %q", test.method, test.path, test.expected, w.Body.String())
			}
		}
	}
}

func Test404NotFound(t *testing.T) {
	app := New()
	app.GET("/exists", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/doesnotexist", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "404") {
		t.Errorf("Expected body to contain '404', got %q", w.Body.String())
	}
}

func TestParameterRoutes(t *testing.T) {
	app := New()
	app.GET("/users/:id", func(c *context.Context) error {
		id := c.Param("id")
		return c.String(200, "User "+id)
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	expected := "User 123"
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("Expected body %q, got %q", expected, w.Body.String())
	}
}

func TestMiddlewareExecution(t *testing.T) {
	app := New()

	// Track middleware execution order
	var order []string

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) error {
			order = append(order, "before1")
			err := next(c)
			order = append(order, "after1")
			return err
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) error {
			order = append(order, "before2")
			err := next(c)
			order = append(order, "after2")
			return err
		}
	}

	app.Use(middleware1, middleware2)

	app.GET("/test", func(c *context.Context) error {
		order = append(order, "handler")
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Middleware should execute in order: m1 -> m2 -> handler -> m2 -> m1
	expected := []string{"before1", "before2", "handler", "after2", "after1"}
	if len(order) != len(expected) {
		t.Fatalf("Expected %d middleware calls, got %d", len(expected), len(order))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("Step %d: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestMiddlewareTermination(t *testing.T) {
	app := New()

	// Middleware that terminates the chain
	authMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) error {
			token := c.Header("Authorization")
			if token != "Bearer valid" {
				return c.JSON(401, map[string]string{"error": "Unauthorized"})
			}
			return next(c)
		}
	}

	app.Use(authMiddleware)
	app.GET("/protected", func(c *context.Context) error {
		return c.String(200, "Secret data")
	})

	// Test without auth
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Test with valid auth
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestErrorHandling(t *testing.T) {
	app := New()

	app.GET("/error", func(c *context.Context) error {
		return errors.New("something went wrong")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("Expected error message in body, got %q", w.Body.String())
	}
}

func TestJSONResponse(t *testing.T) {
	app := New()

	app.GET("/json", func(c *context.Context) error {
		return c.JSON(200, map[string]interface{}{
			"message": "success",
			"code":    200,
		})
	})

	req := httptest.NewRequest("GET", "/json", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
	if !strings.Contains(w.Body.String(), "success") {
		t.Errorf("Expected JSON body to contain 'success', got %q", w.Body.String())
	}
}

func TestStatusMethod(t *testing.T) {
	app := New()

	// Test lazy Status() - should not write headers immediately
	app.GET("/status", func(c *context.Context) error {
		c.Status(201) // Set status
		return c.JSON(201, map[string]string{"created": "yes"})
	})

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestMultipleMiddleware(t *testing.T) {
	app := New()

	// Add multiple middleware
	headerMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) error {
			c.SetHeader("X-Custom", "value")
			return next(c)
		}
	}

	loggingMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) error {
			// In a real scenario, this would log
			return next(c)
		}
	}

	app.Use(headerMiddleware, loggingMiddleware)

	app.GET("/test", func(c *context.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Header().Get("X-Custom") != "value" {
		t.Errorf("Expected X-Custom header to be set")
	}
}

func TestRootPath(t *testing.T) {
	app := New()

	app.GET("/", func(c *context.Context) error {
		return c.String(200, "Home")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != "Home" {
		t.Errorf("Expected body 'Home', got %q", w.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	app := New()

	app.GET("/resource", func(c *context.Context) error {
		return c.String(200, "GET OK")
	})

	// Try POST on a GET-only route
	req := httptest.NewRequest("POST", "/resource", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Should return 404 since POST route doesn't exist
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
