package kese

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, Static File!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	app := New()
	app.StaticFile("/testfile", testFile)

	req := httptest.NewRequest("GET", "/testfile", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if body := strings.TrimSpace(w.Body.String()); body != testContent {
		t.Errorf("Expected body %q, got %q", testContent, body)
	}
}

func TestStaticFileNotFound(t *testing.T) {
	app := New()
	app.StaticFile("/missing", "/nonexistent/file.txt")

	req := httptest.NewRequest("GET", "/missing", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "404") {
		t.Errorf("Expected '404' in response body, got %q", w.Body.String())
	}
}

func TestStaticFileDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	app := New()
	app.StaticFile("/dir", tmpDir)

	req := httptest.NewRequest("GET", "/dir", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Should return 404 when trying to serve a directory
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for directory, got %d", w.Code)
	}
}

func TestStatic(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test files
	testContent := "Static content"
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create subdirectory with file
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	subFile := filepath.Join(subDir, "nested.txt")
	if err := os.WriteFile(subFile, []byte("Nested content"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	app := New()
	app.Static("/assets", tmpDir)

	// Test serving file from root
	req := httptest.NewRequest("GET", "/assets/test.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if body := strings.TrimSpace(w.Body.String()); body != testContent {
		t.Errorf("Expected body %q, got %q", testContent, body)
	}

	// Test serving nested file
	req = httptest.NewRequest("GET", "/assets/sub/nested.txt", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for nested file, got %d", w.Code)
	}
	if body := strings.TrimSpace(w.Body.String()); body != "Nested content" {
		t.Errorf("Expected nested content, got %q", body)
	}
}

func TestStaticNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	app := New()
	app.Static("/assets", tmpDir)

	req := httptest.NewRequest("GET", "/assets/nonexistent.txt", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestStaticIndexHTML(t *testing.T) {
	// Create temporary directory with index.html
	tmpDir := t.TempDir()
	indexContent := "<html><body>Index Page</body></html>"
	indexFile := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(indexFile, []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to create index.html: %v", err)
	}

	app := New()
	app.Static("/public", tmpDir)

	// Test that accessing directory root returns index.html
	tests := []string{"/public", "/public/"}
	for _, path := range tests {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", path, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Index Page") {
			t.Errorf("Expected index.html content for %s, got %q", path, w.Body.String())
		}
	}
}

func TestStaticPathTraversalSecurity(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a file in temp dir
	testFile := filepath.Join(tmpDir, "safe.txt")
	if err := os.WriteFile(testFile, []byte("safe content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to create a file outside tmpDir to test security
	// Note: We're testing that the app prevents access, not that the file exists
	app := New()
	app.Static("/files", tmpDir)

	// Attempt path traversal attacks
	attacks := []string{
		"/files/../../../etc/passwd",
		"/files/../../test",
		"/files/..",
		"/files/../",
	}

	for _, attack := range attacks {
		req := httptest.NewRequest("GET", attack, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		// Should either be forbidden (403) or not found (404), but never succeed
		if w.Code == http.StatusOK {
			t.Errorf("Path traversal attack succeeded: %s returned 200", attack)
		}

		// Most will be 404 because the cleaned path won't exist
		// Some might be 403 if path validation catches them
		if w.Code != http.StatusNotFound && w.Code != http.StatusForbidden {
			t.Errorf("Unexpected status for path traversal %s: got %d", attack, w.Code)
		}
	}
}

func TestStaticMIMETypes(t *testing.T) {
	// Create temporary directory with different file types
	tmpDir := t.TempDir()

	files := map[string]struct {
		content     string
		contentType string
	}{
		"test.txt":  {"text content", "text/plain"},
		"test.html": {"<html></html>", "text/html"},
		"test.css":  {"body{}", "text/css"},
		"test.js":   {"console.log('test')", "text/javascript"},
		"test.json": {"{\"key\":\"value\"}", "application/json"},
	}

	for filename, data := range files {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(data.content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	app := New()
	app.Static("/static", tmpDir)

	for filename, expected := range files {
		req := httptest.NewRequest("GET", "/static/"+filename, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", filename, w.Code)
			continue
		}

		contentType := w.Header().Get("Content-Type")
		// Content-Type might have charset appended
		if !strings.HasPrefix(contentType, expected.contentType) {
			t.Errorf("Expected Content-Type %q for %s, got %q",
				expected.contentType, filename, contentType)
		}
	}
}

func TestStaticWithTrailingSlash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	app := New()
	// Register with trailing slash
	app.Static("/assets/", tmpDir)

	req := httptest.NewRequest("GET", "/assets/file.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with trailing slash in prefix, got %d", w.Code)
	}
}

func TestStaticDirectoryAccess(t *testing.T) {
	// Create directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	app := New()
	app.Static("/files", tmpDir)

	// Try to access the subdirectory directly (should 404)
	req := httptest.NewRequest("GET", "/files/subdir", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for directory access, got %d", w.Code)
	}
}
