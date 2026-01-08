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

	// http.ServeFile redirects directories to add trailing slash (301),
	// which is standard HTTP behavior
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status 301 for directory redirect, got %d", w.Code)
	}
}

func TestStaticSingleLevel(t *testing.T) {
	// Create temporary directory with files
	tmpDir := t.TempDir()

	testContent := "Static content"
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	app := New()
	app.Static("/assets", tmpDir)

	// Test serving file from root level
	req := httptest.NewRequest("GET", "/assets/test.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if body := strings.TrimSpace(w.Body.String()); body != testContent {
		t.Errorf("Expected body %q, got %q", testContent, body)
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
		"test.js":   {"console.log('test')", "application/javascript"},
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
	app.Static("/assets/", tmpDir)

	req := httptest.NewRequest("GET", "/assets/file.txt", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with trailing slash in prefix, got %d", w.Code)
	}
}

func TestMultipleStaticFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	app := New()
	app.StaticFile("/static1", file1)
	app.StaticFile("/static2", file2)

	req := httptest.NewRequest("GET", "/static1", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for file1, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/static2", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for file2, got %d", w.Code)
	}
}
