package middleware

import (
	"fmt"
	"time"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/ratelimit"
)

// RateLimitConfig holds configuration for rate limiting middleware.
type RateLimitConfig struct {
	// Limit is the maximum number of requests allowed in the window
	Limit int

	// Window is the time window for rate limiting
	Window time.Duration

	// KeyFunc generates the rate limit key from the context.
	// Default: uses client IP address
	KeyFunc func(*context.Context) string

	// Store is the storage backend for rate limiting.
	// Default: in-memory store
	Store ratelimit.Store

	// SkipFunc allows skipping rate limiting for certain requests.
	// Return true to skip rate limiting for this request.
	SkipFunc func(*context.Context) bool

	// Message is the error message returned when rate limit is exceeded.
	// Default: "rate limit exceeded"
	Message string
}

// DefaultRateLimitConfig returns the default rate limit configuration.
func DefaultRateLimitConfig(limit int, window time.Duration) RateLimitConfig {
	return RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFunc: func(c *context.Context) string {
			// Use X-Forwarded-For if available, otherwise use RemoteAddr
			if forwarded := c.Header("X-Forwarded-For"); forwarded != "" {
				return forwarded
			}
			return c.Request.RemoteAddr
		},
		Store:    ratelimit.NewMemoryStore(),
		SkipFunc: nil,
		Message:  "rate limit exceeded",
	}
}

// RateLimit returns a middleware that limits requests based on IP address.
//
// limit: Maximum number of requests allowed
// window: Time window for the limit (e.g., time.Minute, time.Hour)
//
// Example:
//
//	// 100 requests per minute
//	app.Use(middleware.RateLimit(100, time.Minute))
func RateLimit(limit int, window time.Duration) kese.MiddlewareFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig(limit, window))
}

// RateLimitWithConfig returns a rate limiting middleware with custom configuration.
//
// Example:
//
//	app.Use(middleware.RateLimitWithConfig(RateLimitConfig{
//	    Limit: 1000,
//	    Window: time.Hour,
//	    KeyFunc: func(c *context.Context) string {
//	        // Rate limit per user instead of IP
//	        user := c.Get("userID")
//	        if user != nil {
//	            return fmt.Sprintf("user:%v", user)
//	        }
//	        return c.Request.RemoteAddr
//	    },
//	}))
func RateLimitWithConfig(config RateLimitConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Check if we should skip rate limiting
			if config.SkipFunc != nil && config.SkipFunc(c) {
				return next(c)
			}

			// Get rate limit key
			key := config.KeyFunc(c)

			// Increment counter
			count, err := config.Store.Increment(key, config.Window)
			if err != nil {
				// On error, allow the request but log it
				fmt.Printf("Rate limit error: %v\n", err)
				return next(c)
			}

			// Set rate limit headers
			c.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
			c.SetHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, config.Limit-count)))

			// Check if limit exceeded
			if count > config.Limit {
				c.SetHeader("Retry-After", fmt.Sprintf("%d", int(config.Window.Seconds())))
				return c.JSON(429, map[string]string{
					"error": config.Message,
				})
			}

			return next(c)
		}
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
