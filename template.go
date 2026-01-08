package kese

import (
	"bytes"
	"html/template"
	"path/filepath"
	"sync"

	"github.com/JedizLaPulga/kese/context"
)

// TemplateEngine manages HTML template rendering.
type TemplateEngine struct {
	templates *template.Template
	dir       string
	mu        sync.RWMutex
}

// NewTemplateEngine creates a new template engine with the given directory.
func NewTemplateEngine(dir string) *TemplateEngine {
	return &TemplateEngine{
		dir: dir,
	}
}

// LoadTemplates loads all templates from the template directory.
// Supports glob patterns like "*.html" or "**/*.html"
func (te *TemplateEngine) LoadTemplates(pattern string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	// Parse all templates matching the pattern
	tmpl, err := template.ParseGlob(filepath.Join(te.dir, pattern))
	if err != nil {
		return err
	}

	te.templates = tmpl
	return nil
}

// Render renders a template with the given data and writes it using the context.
// The output is buffered to prevent partial rendering if an error occurs.
func (te *TemplateEngine) Render(c *context.Context, status int, name string, data interface{}) error {
	te.mu.RLock()
	defer te.mu.RUnlock()

	if te.templates == nil {
		return c.InternalError("Templates not loaded")
	}

	// Buffer the template output
	var buf bytes.Buffer
	err := te.templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		// Template execution failed - return error without writing partial response
		return err
	}

	// Template executed successfully - write the complete response
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	c.Writer.Write(buf.Bytes())
	c.SetWritten()
	return nil
}
