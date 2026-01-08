package context

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

// TestBodyMultipleReads verifies that Body() can be called multiple times
func TestBodyMultipleReads(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	body := `{"name":"John","email":"john@example.com"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/users", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")

	ctx := New(w, r)

	// First read
	var user1 User
	err := ctx.Body(&user1)
	if err != nil {
		t.Fatalf("First Body() call error: %v", err)
	}

	// Second read - should work now with buffering
	var user2 User
	err = ctx.Body(&user2)
	if err != nil {
		t.Fatalf("Second Body() call error: %v", err)
	}

	// Verify both reads got the same data
	if user1.Name != user2.Name || user1.Email != user2.Email {
		t.Error("Multiple Body() reads didn't return the same data")
	}

	if user1.Name != "John" || user1.Email != "john@example.com" {
		t.Error("Body() didn't parse JSON correctly")
	}
}

// TestBodyBytesMultipleReads verifies that BodyBytes() can be called multiple times
func TestBodyBytesMultipleReads(t *testing.T) {
	body := "test body content"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))

	ctx := New(w, r)

	// First read
	data1, err := ctx.BodyBytes()
	if err != nil {
		t.Fatalf("First BodyBytes() call error: %v", err)
	}

	// Second read - should work now with buffering
	data2, err := ctx.BodyBytes()
	if err != nil {
		t.Fatalf("Second BodyBytes() call error: %v", err)
	}

	// Verify both reads got the same data
	if string(data1) != string(data2) {
		t.Error("Multiple BodyBytes() reads didn't return the same data")
	}

	if string(data1) != body {
		t.Errorf("Expected body=%s, got %s", body, string(data1))
	}
}

// TestBodyAndBodyBytesMixed verifies that Body() and BodyBytes() can be used together
func TestBodyAndBodyBytesMixed(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}

	body := `{"name":"Alice"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/users", bytes.NewBufferString(body))

	ctx := New(w, r)

	// First read with BodyBytes
	data, err := ctx.BodyBytes()
	if err != nil {
		t.Fatalf("BodyBytes() error: %v", err)
	}

	// Second read with Body() - should work with buffered data
	var user User
	err = ctx.Body(&user)
	if err != nil {
		t.Fatalf("Body() error after BodyBytes(): %v", err)
	}

	if string(data) != body {
		t.Error("BodyBytes() didn't return correct data")
	}

	if user.Name != "Alice" {
		t.Error("Body() didn't parse JSON after BodyBytes() call")
	}
}

// TestQueryDefaultWithEmptyValue verifies QueryDefault distinguishes missing vs empty
func TestQueryDefaultWithEmptyValue(t *testing.T) {
	w := httptest.NewRecorder()

	// Test with empty value: ?q=
	r1 := httptest.NewRequest("GET", "/search?q=", nil)
	ctx1 := New(w, r1)

	result1 := ctx1.QueryDefault("q", "default")
	if result1 != "" {
		t.Errorf("Expected empty string for ?q=, got %q", result1)
	}

	// Test with missing parameter
	r2 := httptest.NewRequest("GET", "/search", nil)
	ctx2 := New(w, r2)

	result2 := ctx2.QueryDefault("q", "default")
	if result2 != "default" {
		t.Errorf("Expected 'default' for missing param, got %q", result2)
	}

	// Test with actual value
	r3 := httptest.NewRequest("GET", "/search?q=test", nil)
	ctx3 := New(w, r3)

	result3 := ctx3.QueryDefault("q", "default")
	if result3 != "test" {
		t.Errorf("Expected 'test', got %q", result3)
	}
}
