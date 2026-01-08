package middleware

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/logger"
)

// Logger returns a middleware that logs HTTP requests using structured logging.
// It logs the method, path, status code, and response time for each request.
// Accepts a logger instance to ensure consistent structured logging across the application.
func Logger(logger *logger.Logger) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			start := time.Now()

			// Call the next handler
			err := next(c)

			// Log after handler completes using structured logging
			duration := time.Since(start)
			logger.Info("Request completed",
				"method", c.Method(),
				"path", c.Path(),
				"status", c.StatusCode(),
				"duration_ms", duration.Milliseconds(),
			)

			return err
		}
	}
}

// Recovery returns a middleware that recovers from panics using structured logging.
// It prevents the server from crashing and returns a 500 error.
// Accepts a logger instance to ensure panic details are logged with proper structure.
func Recovery(logger *logger.Logger) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Log panic with structured logging
					logger.Error("Panic recovered",
						"panic", fmt.Sprintf("%v", r),
						"stack", string(debug.Stack()),
					)
					// Only write response if nothing has been written yet
					if !c.IsWritten() {
						c.JSON(500, map[string]interface{}{
							"error": "Internal Server Error",
						})
					}
				}
			}()

			return next(c)
		}
	}
}

// CORS returns a middleware that adds CORS headers to responses.
// This allows cross-origin requests from web browsers.
func CORS() kese.MiddlewareFunc {
	return CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	})
}

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// CORSWithConfig returns a CORS middleware with custom configuration.
// Properly handles multiple allowed origins by checking the request origin.
func CORSWithConfig(config CORSConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Set CORS headers based on configuration
			if len(config.AllowOrigins) > 0 {
				// Check if wildcard is allowed
				if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
					c.SetHeader("Access-Control-Allow-Origin", "*")
				} else {
					// Match request origin against allowed origins
					requestOrigin := c.Header("Origin")
					for _, allowedOrigin := range config.AllowOrigins {
						if allowedOrigin == requestOrigin {
							c.SetHeader("Access-Control-Allow-Origin", requestOrigin)
							c.SetHeader("Access-Control-Allow-Credentials", "true")
							break
						}
					}
				}
			}

			if len(config.AllowMethods) > 0 {
				c.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			}

			if len(config.AllowHeaders) > 0 {
				c.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			}

			// Handle preflight requests
			if c.Method() == "OPTIONS" {
				c.NoContent()
				return nil
			}

			return next(c)
		}
	}
}

// RequestID returns a middleware that adds a unique request ID to each request.
// The ID is set in the X-Request-ID header.
// Uses atomic operations to safely increment the counter across concurrent requests.
func RequestID() kese.MiddlewareFunc {
	var counter atomic.Uint64

	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			count := counter.Add(1)
			requestID := fmt.Sprintf("%d-%d", time.Now().Unix(), count)
			c.SetHeader("X-Request-ID", requestID)
			return next(c)
		}
	}
}
