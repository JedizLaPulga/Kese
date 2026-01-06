package middleware

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
)

// Logger returns a middleware that logs HTTP requests.
// It logs the method, path, and response time for each request.
func Logger() kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			start := time.Now()

			// Call the next handler
			err := next(c)

			// Log after handler completes
			duration := time.Since(start)
			log.Printf("[%s] %s - %d - %v",
				c.Method(),
				c.Path(),
				c.StatusCode(),
				duration,
			)

			return err
		}
	}
}

// Recovery returns a middleware that recovers from panics.
// It prevents the server from crashing and returns a 500 error.
func Recovery() kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC: %v\n%s", r, debug.Stack())
					c.Status(500)
					c.JSON(500, map[string]interface{}{
						"error": "Internal Server Error",
						"panic": fmt.Sprintf("%v", r),
					})
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
func CORSWithConfig(config CORSConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Set CORS headers
			if len(config.AllowOrigins) > 0 {
				origin := config.AllowOrigins[0]
				if origin == "*" {
					c.SetHeader("Access-Control-Allow-Origin", "*")
				} else {
					c.SetHeader("Access-Control-Allow-Origin", origin)
				}
			}

			if len(config.AllowMethods) > 0 {
				methods := ""
				for i, method := range config.AllowMethods {
					if i > 0 {
						methods += ", "
					}
					methods += method
				}
				c.SetHeader("Access-Control-Allow-Methods", methods)
			}

			if len(config.AllowHeaders) > 0 {
				headers := ""
				for i, header := range config.AllowHeaders {
					if i > 0 {
						headers += ", "
					}
					headers += header
				}
				c.SetHeader("Access-Control-Allow-Headers", headers)
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
