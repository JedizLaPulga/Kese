package kese

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/JedizLaPulga/kese/context"
)

// Static serves files from a directory at the given URL path prefix.
// Example: app.Static("/assets", "./public") serves ./public/style.css at /assets/style.css
// Note: Currently only supports single-level paths (e.g., /assets/file.css)
// Nested paths (e.g., /assets/sub/file.css) are not supported due to router design
func (a *App) Static(urlPrefix, fsPath string) {
	// Normalize the URL prefix
	urlPrefix = strings.TrimSuffix(urlPrefix, "/")

	handler := func(c *context.Context) error {
		// Get the requested filename from the :filepath parameter
		filename := c.Param("filepath")

		// If no filename provided, return 404
		if filename == "" {
			return c.String(http.StatusNotFound, "404 Not Found")
		}

		// Build the full file path
		filePath := filepath.Join(fsPath, filepath.Clean(filename))

		// Security check: ensure the file is within fsPath
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		absFsPath, err := filepath.Abs(fsPath)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		if !strings.HasPrefix(absPath+string(filepath.Separator), absFsPath+string(filepath.Separator)) &&
			absPath != absFsPath {
			return c.String(http.StatusForbidden, "Forbidden")
		}

		// Serve the file - http.ServeFile handles existence checks, MIME types, caching, etc.
		http.ServeFile(c.Writer, c.Request, filePath)
		c.SetWritten()
		return nil
	}

	// Register a parameter-based route for this prefix
	a.GET(urlPrefix+"/:filepath", handler)
}

// StaticFile serves a single file at the given URL path.
// Example: app.StaticFile("/favicon.ico", "./assets/favicon.ico")
func (a *App) StaticFile(urlPath, filePath string) {
	handler := func(c *context.Context) error {
		// Serve the file - http.ServeFile handles existence checks, directories, etc.
		http.ServeFile(c.Writer, c.Request, filePath)
		c.SetWritten()
		return nil
	}

	a.GET(urlPath, handler)
}
