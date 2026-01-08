package router

import (
	"testing"
)

// Use string as the handler type for testing simplicity
const mockHandler = "mockHandler"

func TestNew(t *testing.T) {
	r := New[string]()
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.trees == nil {
		t.Fatal("Router trees not initialized")
	}
}

func TestAddStaticRoute(t *testing.T) {
	r := New[string]()
	r.Add("GET", "/users", mockHandler)

	handler, params, found := r.Match("GET", "/users")
	if !found || handler == "" {
		t.Fatal("Handler should not be empty")
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %d", len(params))
	}
}

func TestAddRootRoute(t *testing.T) {
	r := New[string]()
	r.Add("GET", "/", mockHandler)

	handler, params, found := r.Match("GET", "/")
	if !found || handler == "" {
		t.Fatal("Handler should not be empty")
	}
	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %d", len(params))
	}
}

func TestAddParameterRoute(t *testing.T) {
	r := New[string]()
	r.Add("GET", "/users/:id", mockHandler)

	handler, params, found := r.Match("GET", "/users/123")
	if !found || handler == "" {
		t.Fatal("Handler should not be empty")
	}
	if len(params) != 1 {
		t.Fatalf("Expected 1 param, got %d", len(params))
	}
	if params.Get("id") != "123" {
		t.Errorf("Expected param id=123, got %s", params.Get("id"))
	}
}

func TestMultipleParameters(t *testing.T) {
	r := New[string]()
	r.Add("GET", "/users/:userId/posts/:postId", mockHandler)

	handler, params, found := r.Match("GET", "/users/42/posts/100")
	if !found || handler == "" {
		t.Fatal("Handler should not be empty")
	}
	if len(params) != 2 {
		t.Fatalf("Expected 2 params, got %d", len(params))
	}
	if params.Get("userId") != "42" {
		t.Errorf("Expected userId=42, got %s", params.Get("userId"))
	}
	if params.Get("postId") != "100" {
		t.Errorf("Expected postId=100, got %s", params.Get("postId"))
	}
}

func TestMixedStaticAndParams(t *testing.T) {
	r := New[string]()
	r.Add("GET", "/api/v1/users/:id/profile", mockHandler)

	handler, params, found := r.Match("GET", "/api/v1/users/999/profile")
	if !found || handler == "" {
		t.Fatal("Handler should not be empty")
	}
	if params.Get("id") != "999" {
		t.Errorf("Expected id=999, got %s", params.Get("id"))
	}
}

func TestMethodSeparation(t *testing.T) {
	r := New[string]()
	getHandler := "getHandler"
	postHandler := "postHandler"
	r.Add("GET", "/users", getHandler)
	r.Add("POST", "/users", postHandler)

	// Both methods should work independently
	handler, _, found := r.Match("GET", "/users")
	if !found || handler != getHandler {
		t.Fatal("GET handler match failed")
	}

	handler, _, found = r.Match("POST", "/users")
	if !found || handler != postHandler {
		t.Fatal("POST handler match failed")
	}

	// Non-existent method should return nil
	_, _, found = r.Match("DELETE", "/users")
	if found {
		t.Fatal("DELETE handler should be nil/not found")
	}
}

func TestNoMatch(t *testing.T) {
	r := New[string]()
	r.Add("GET", "/users", mockHandler)

	_, _, found := r.Match("GET", "/posts")
	if found {
		t.Fatal("Handler should be nil/not found for non-matching route")
	}
}

func TestTrailingSlash(t *testing.T) {
	r := New[string]()
	r.Add("GET", "/users", mockHandler)

	// Without trailing slash
	handler, _, found := r.Match("GET", "/users")
	if !found || handler == "" {
		t.Fatal("Should match /users")
	}

	// With trailing slash - should also match because splitPath normalizes it
	handler, _, found = r.Match("GET", "/users/")
	if !found || handler == "" {
		t.Fatal("Should match /users/ when route is /users (trailing slash is normalized)")
	}

}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/", []string{}},
		{"/users", []string{"users"}},
		{"/users/", []string{"users"}},
		{"/users/123", []string{"users", "123"}},
		{"/api/v1/users/:id", []string{"api", "v1", "users", ":id"}},
		{"//users//123//", []string{"users", "123"}},
	}

	for _, test := range tests {
		result := splitPath(test.path)
		if len(result) != len(test.expected) {
			t.Errorf("splitPath(%q): expected %d segments, got %d", test.path, len(test.expected), len(result))
			continue
		}
		for i, segment := range result {
			if segment != test.expected[i] {
				t.Errorf("splitPath(%q): expected segment %d to be %q, got %q", test.path, i, test.expected[i], segment)
			}
		}
	}
}

func TestComplexRouting(t *testing.T) {
	r := New[string]()

	// Register multiple routes
	r.Add("GET", "/", mockHandler)
	r.Add("GET", "/users", mockHandler)
	r.Add("GET", "/users/:id", mockHandler)
	r.Add("POST", "/users", mockHandler)
	r.Add("GET", "/users/:id/posts", mockHandler)
	r.Add("GET", "/users/:id/posts/:postId", mockHandler)

	tests := []struct {
		method         string
		path           string
		shouldMatch    bool
		expectedParams map[string]string
	}{
		{"GET", "/", true, map[string]string{}},
		{"GET", "/users", true, map[string]string{}},
		{"GET", "/users/42", true, map[string]string{"id": "42"}},
		{"POST", "/users", true, map[string]string{}},
		{"GET", "/users/42/posts", true, map[string]string{"id": "42"}},
		{"GET", "/users/42/posts/100", true, map[string]string{"id": "42", "postId": "100"}},
		{"DELETE", "/users", false, nil},
		{"GET", "/posts", false, nil},
		{"GET", "/users/42/comments", false, nil},
	}

	for _, test := range tests {
		handler, params, found := r.Match(test.method, test.path)

		if test.shouldMatch {
			if !found || handler == "" {
				t.Errorf("%s %s: expected match, got nil handler", test.method, test.path)
			}
			if len(params) != len(test.expectedParams) {
				t.Errorf("%s %s: expected %d params, got %d", test.method, test.path, len(test.expectedParams), len(params))
			}
			for key, expectedValue := range test.expectedParams {
				if params.Get(key) != expectedValue {
					t.Errorf("%s %s: expected param %s=%s, got %s", test.method, test.path, key, expectedValue, params.Get(key))
				}
			}
		} else {
			if found {
				t.Errorf("%s %s: expected no match, got handler", test.method, test.path)
			}
		}
	}
}
