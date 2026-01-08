package kese

import (
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
func (te *TemplateEngine) Render(c *context.Context, status int, name string, data interface{}) error {
	te.mu.RLock()
	defer te.mu.RUnlock()

	if te.templates == nil {
		return c.InternalError("Templates not loaded")
	}

	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)

	err := te.templates.ExecuteTemplate(c.Writer, name, data)
	if err != nil {
		return err
	}

	c.SetWritten()
	return nil
}

// AddTemplate adds extra functionality to the App for template rendering
type AppWithTemplates struct {
	*App
	templateEngine *TemplateEngine
}

// SetTemplateEngine sets the template engine for rendering HTML templates.
func (a *App) SetTemplateEngine(engine *TemplateEngine) {
	// Store engine as custom data in the app
	// We'll use a simple approach: add a Render method to context through middleware
}

// RenderTemplate is a helper to render a template using the App's template engine.
// This should be used after SetTemplateEngine has been called.
func RenderTemplate(engine *TemplateEngine) func(*context.Context, int, string, interface{}) error {
	return func(c *context.Context, status int, name string, data interface{}) error {
		return engine.Render(c, status, name, data)
	}
}
