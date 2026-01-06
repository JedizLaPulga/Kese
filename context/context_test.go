package context

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)
	if ctx == nil {
		t.Fatal("New() returned nil")
	}
	if ctx.Request != r {
		t.Error("Context.Request not set correctly")
	}
	if ctx.Writer != w {
		t.Error("Context.Writer not set correctly")
	}
	if ctx.params == nil {
		t.Error("Context.params not initialized")
	}
}

func TestParam(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/123", nil)

	ctx := New(w, r)
	ctx.SetParams(map[string]string{"id": "123", "name": "john"})

	if ctx.Param("id") != "123" {
		t.Errorf("Expected id=123, got %s", ctx.Param("id"))
	}
	if ctx.Param("name") != "john" {
		t.Errorf("Expected name=john, got %s", ctx.Param("name"))
	}
	if ctx.Param("nonexistent") != "" {
		t.Error("Expected empty string for non-existent param")
	}
}

func TestQuery(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q=golang&page=2", nil)

	ctx := New(w, r)

	if ctx.Query("q") != "golang" {
		t.Errorf("Expected q=golang, got %s", ctx.Query("q"))
	}
	if ctx.Query("page") != "2" {
		t.Errorf("Expected page=2, got %s", ctx.Query("page"))
	}
	if ctx.Query("nonexistent") != "" {
		t.Error("Expected empty string for non-existent query param")
	}
}

func TestQueryDefault(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q=golang", nil)

	ctx := New(w, r)

	if ctx.QueryDefault("q", "default") != "golang" {
		t.Error("Should return actual value when present")
	}
	if ctx.QueryDefault("page", "1") != "1" {
		t.Error("Should return default value when not present")
	}
}

func TestHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer token123")
	r.Header.Set("Content-Type", "application/json")

	ctx := New(w, r)

	if ctx.Header("Authorization") != "Bearer token123" {
		t.Error("Failed to get Authorization header")
	}
	if ctx.Header("Content-Type") != "application/json" {
		t.Error("Failed to get Content-Type header")
	}
}

func TestSetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)
	ctx.SetHeader("X-Custom-Header", "test-value")

	if w.Header().Get("X-Custom-Header") != "test-value" {
		t.Error("Failed to set response header")
	}
}

func TestStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)
	ctx.Status(http.StatusCreated)

	// Status() is now lazy - it doesn't write headers immediately
	// It only sets the internal statusCode
	if ctx.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() returned %d, expected 201", ctx.StatusCode())
	}

	// The writer should still have default status until a response is written
	if w.Code != http.StatusOK {
		t.Errorf("Expected writer status 200 (default), got %d", w.Code)
	}

	// Now write a response - it should use the stored status code
	ctx.String(ctx.StatusCode(), "Created")

	// Now the writer should have the correct status
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 after response, got %d", w.Code)
	}
}

func TestBody(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	body := `{"name":"John","email":"john@example.com"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/users", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")

	ctx := New(w, r)

	var user User
	err := ctx.Body(&user)
	if err != nil {
		t.Fatalf("Body() error: %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Expected name=John, got %s", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email=john@example.com, got %s", user.Email)
	}
}

func TestBodyBytes(t *testing.T) {
	body := "test body content"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))

	ctx := New(w, r)

	data, err := ctx.BodyBytes()
	if err != nil {
		t.Fatalf("BodyBytes() error: %v", err)
	}

	if string(data) != body {
		t.Errorf("Expected body=%s, got %s", body, string(data))
	}
}

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)

	data := map[string]interface{}{
		"message": "Hello",
		"code":    200,
	}

	err := ctx.JSON(http.StatusOK, data)
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type=application/json, got %s", contentType)
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Validate JSON content
	var result map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["message"] != "Hello" {
		t.Error("JSON response doesn't contain expected data")
	}
}

func TestString(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)

	err := ctx.String(http.StatusOK, "Hello, World!")
	if err != nil {
		t.Fatalf("String() error: %v", err)
	}

	if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Error("Content-Type not set correctly")
	}
	if w.Body.String() != "Hello, World!" {
		t.Errorf("Expected body='Hello, World!', got %s", w.Body.String())
	}
}

func TestHTML(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)

	html := "<h1>Hello</h1>"
	err := ctx.HTML(http.StatusOK, html)
	if err != nil {
		t.Fatalf("HTML() error: %v", err)
	}

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("Content-Type not set correctly")
	}
	if w.Body.String() != html {
		t.Errorf("Expected body=%s, got %s", html, w.Body.String())
	}
}

func TestBytes(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)

	data := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f} // "Hello"
	err := ctx.Bytes(http.StatusOK, "application/octet-stream", data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}

	if w.Header().Get("Content-Type") != "application/octet-stream" {
		t.Error("Content-Type not set correctly")
	}
	if !bytes.Equal(w.Body.Bytes(), data) {
		t.Error("Body bytes don't match")
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/users/123", nil)

	ctx := New(w, r)

	err := ctx.NoContent()
	if err != nil {
		t.Fatalf("NoContent() error: %v", err)
	}

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Error("Expected empty body")
	}
}

func TestRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/old", nil)

	ctx := New(w, r)

	err := ctx.Redirect(http.StatusMovedPermanently, "/new")
	if err != nil {
		t.Fatalf("Redirect() error: %v", err)
	}

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status 301, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/new" {
		t.Error("Location header not set correctly")
	}
}

func TestRedirectInvalidStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/old", nil)

	ctx := New(w, r)

	err := ctx.Redirect(http.StatusOK, "/new")
	if err == nil {
		t.Error("Expected error for invalid redirect status")
	}
}

func TestCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})

	ctx := New(w, r)

	cookie, err := ctx.Cookie("session")
	if err != nil {
		t.Fatalf("Cookie() error: %v", err)
	}
	if cookie.Value != "abc123" {
		t.Errorf("Expected cookie value=abc123, got %s", cookie.Value)
	}
}

func TestSetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)

	cookie := &http.Cookie{
		Name:  "token",
		Value: "xyz789",
		Path:  "/",
	}
	ctx.SetCookie(cookie)

	// Check if cookie is in response
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "token" || cookies[0].Value != "xyz789" {
		t.Error("Cookie not set correctly")
	}
}

func TestMethod(t *testing.T) {
	tests := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, "/", nil)
		ctx := New(w, r)

		if ctx.Method() != method {
			t.Errorf("Expected method %s, got %s", method, ctx.Method())
		}
	}
}

func TestPath(t *testing.T) {
	paths := []string{"/", "/users", "/users/123", "/api/v1/posts"}

	for _, path := range paths {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", path, nil)
		ctx := New(w, r)

		if ctx.Path() != path {
			t.Errorf("Expected path %s, got %s", path, ctx.Path())
		}
	}
}

func TestIsWritten(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := New(w, r)

	if ctx.IsWritten() {
		t.Error("IsWritten should be false initially")
	}

	ctx.String(200, "test")

	if !ctx.IsWritten() {
		t.Error("IsWritten should be true after writing response")
	}
}
