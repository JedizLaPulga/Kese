package middleware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/cache"
	"github.com/JedizLaPulga/kese/context"
)

// CacheConfig holds configuration for cache middleware.
type CacheConfig struct {
	// TTL is the time-to-live for cached responses
	TTL time.Duration

	// Store is the cache storage backend
	Store cache.Store

	// KeyFunc generates cache keys from context
	// Default: uses method + path
	KeyFunc func(*context.Context) string
}

// DefaultCacheConfig returns default cache configuration.
func DefaultCacheConfig(ttl time.Duration) CacheConfig {
	return CacheConfig{
		TTL:   ttl,
		Store: cache.NewMemoryStore(),
		KeyFunc: func(c *context.Context) string {
			return c.Method() + ":" + c.Path()
		},
	}
}

// Cache returns a middleware that caches GET responses.
//
// Example:
//
//	app.Use(middleware.Cache(5 * time.Minute))
func Cache(ttl time.Duration) kese.MiddlewareFunc {
	return CacheWithConfig(DefaultCacheConfig(ttl))
}

// CacheWithConfig returns cache middleware with custom configuration.
func CacheWithConfig(config CacheConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Only cache GET requests
			if c.Method() != "GET" {
				return next(c)
			}

			// Generate cache key
			key := config.KeyFunc(c)

			// Try to get from cache
			if cached, found := config.Store.Get(key); found {
				// Serve from cache
				c.SetHeader("X-Cache", "HIT")
				c.SetHeader("Content-Type", "application/json")
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write(cached)
				c.SetWritten()
				return nil
			}

			// Capture response
			recorder := &responseRecorder{
				ResponseWriter: c.Writer,
				body:           &bytes.Buffer{},
			}

			c.Writer = recorder

			// Call next handler
			err := next(c)

			// Cache the response if successful
			if err == nil && recorder.statusCode >= 200 && recorder.statusCode < 300 {
				config.Store.Set(key, recorder.body.Bytes(), config.TTL)
			}

			// Set cache miss header
			recorder.Header().Set("X-Cache", "MISS")

			// Write the captured response
			for k, v := range recorder.Header() {
				c.Writer.Header()[k] = v
			}
			if recorder.statusCode > 0 {
				c.Writer.WriteHeader(recorder.statusCode)
			}
			c.Writer.Write(recorder.body.Bytes())

			return err
		}
	}
}

// responseRecorder captures the response for caching.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

// generateCacheKey creates a cache key from request.
func generateCacheKey(method, path, query string) string {
	h := sha256.New()
	io.WriteString(h, method)
	io.WriteString(h, path)
	io.WriteString(h, query)
	return fmt.Sprintf("%x", h.Sum(nil))
}
