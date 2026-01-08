package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
)

// GzipConfig holds configuration for gzip compression middleware.
type GzipConfig struct {
	// Level is the compression level (0-9). Default: gzip.DefaultCompression (-1)
	// 0 = no compression, 1 = best speed, 9 = best compression
	Level int

	// MinLength is the minimum response size to compress in bytes.
	// Responses smaller than this won't be compressed. Default: 1024 (1KB)
	MinLength int64

	// ExcludedExtensions are file extensions that should not be compressed.
	// Default: images and compressed files (.jpg, .png, .gif, .zip, .pdf, etc.)
	ExcludedExtensions []string

	// ExcludedPaths are URL paths that should not be compressed.
	ExcludedPaths []string
}

// DefaultGzipConfig returns the default gzip configuration.
func DefaultGzipConfig() GzipConfig {
	return GzipConfig{
		Level:     gzip.DefaultCompression, //-1
		MinLength: 1024,                    // 1KB
		ExcludedExtensions: []string{
			".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico",
			".zip", ".gz", ".tar", ".rar", ".7z",
			".mp4", ".avi", ".mov", ".mp3", ".wav",
			".pdf",
		},
		ExcludedPaths: []string{},
	}
}

// gzipResponseWriter wraps http.ResponseWriter to enable gzip compression.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	wroteHeader bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.ResponseWriter.Header().Del("Content-Length") // Content-Length will be wrong after compression
	w.ResponseWriter.WriteHeader(code)
}

// Gzip returns a middleware that compresses HTTP responses using gzip.
// Uses default configuration (compression level -1, min size 1KB).
//
// Example:
//
//	app.Use(middleware.Gzip())
func Gzip() kese.MiddlewareFunc {
	return GzipWithConfig(DefaultGzipConfig())
}

// GzipWithConfig returns a gzip middleware with custom configuration.
//
// Example:
//
//	app.Use(middleware.GzipWithConfig(GzipConfig{
//	    Level: 6,  // Higher compression
//	    MinLength: 512, // Compress files > 512 bytes
//	}))
func GzipWithConfig(config GzipConfig) kese.MiddlewareFunc {
	// Build excluded extensions map for fast lookup
	excludedExts := make(map[string]bool)
	for _, ext := range config.ExcludedExtensions {
		excludedExts[ext] = true
	}

	// Build excluded paths map
	excludedPaths := make(map[string]bool)
	for _, path := range config.ExcludedPaths {
		excludedPaths[path] = true
	}

	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Check if client accepts gzip
			if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
				return next(c)
			}

			// Check if path is excluded
			if excludedPaths[c.Request.URL.Path] {
				return next(c)
			}

			// Check if file extension is excluded
			path := c.Request.URL.Path
			for ext := range excludedExts {
				if strings.HasSuffix(path, ext) {
					return next(c)
				}
			}

			// Create gzip writer
			gz, err := gzip.NewWriterLevel(c.Writer, config.Level)
			if err != nil {
				return next(c) // Fall back to no compression
			}
			defer gz.Close()

			// Set gzip headers
			c.Writer.Header().Set("Content-Encoding", "gzip")
			c.Writer.Header().Set("Vary", "Accept-Encoding")

			// Wrap response writer
			gzWriter := &gzipResponseWriter{
				Writer:         gz,
				ResponseWriter: c.Writer,
			}

			// Replace the writer in context
			originalWriter := c.Writer
			c.Writer = gzWriter

			// Call next handler
			err = next(c)

			// Restore original writer
			c.Writer = originalWriter

			return err
		}
	}
}
