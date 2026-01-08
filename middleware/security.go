package middleware

import (
	"fmt"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
)

// SecurityConfig holds configuration for security headers middleware.
type SecurityConfig struct {
	// XFrameOptions controls iframe embedding. Default: "DENY"
	XFrameOptions string

	// ContentTypeNosniff prevents MIME type sniffing. Default: "nosniff"
	ContentTypeNosniff string

	// XSSProtection enables browser XSS protection. Default: "1; mode=block"
	XSSProtection string

	// HSTSMaxAge is the max age for Strict-Transport-Security header in seconds.
	// Set to 0 to disable HSTS. Default: 31536000 (1 year)
	HSTSMaxAge int

	// HSTSIncludeSubdomains includes subdomains in HSTS. Default: false
	HSTSIncludeSubdomains bool

	// ContentSecurityPolicy sets CSP header. Empty string disables CSP.
	ContentSecurityPolicy string

	// ReferrerPolicy controls Referer header. Default: "strict-origin-when-cross-origin"
	ReferrerPolicy string
}

// DefaultSecurityConfig returns the default security configuration.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		XFrameOptions:         "DENY",
		ContentTypeNosniff:    "nosniff",
		XSSProtection:         "1; mode=block",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: false,
		ContentSecurityPolicy: "",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// SecureHeaders returns a middleware that adds security headers to all responses.
// Uses default security configuration.
//
// Headers added:
//   - X-Frame-Options: DENY
//   - X-Content-Type-Options: nosniff
//   - X-XSS-Protection: 1; mode=block
//   - Strict-Transport-Security: max-age=31536000
//   - Referrer-Policy: strict-origin-when-cross-origin
//
// Example:
//
//	app.Use(middleware.SecureHeaders())
func SecureHeaders() kese.MiddlewareFunc {
	return SecureHeadersWithConfig(DefaultSecurityConfig())
}

// SecureHeadersWithConfig returns a middleware with custom security configuration.
//
// Example:
//
//	app.Use(middleware.SecureHeadersWithConfig(SecurityConfig{
//	    XFrameOptions: "SAMEORIGIN",
//	    HSTSMaxAge: 63072000, // 2 years
//	    ContentSecurityPolicy: "default-src 'self'",
//	}))
func SecureHeadersWithConfig(config SecurityConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// X-Frame-Options: prevents clickjacking
			if config.XFrameOptions != "" {
				c.SetHeader("X-Frame-Options", config.XFrameOptions)
			}

			// X-Content-Type-Options: prevents MIME sniffing
			if config.ContentTypeNosniff != "" {
				c.SetHeader("X-Content-Type-Options", config.ContentTypeNosniff)
			}

			// X-XSS-Protection: enables browser XSS filtering
			if config.XSSProtection != "" {
				c.SetHeader("X-XSS-Protection", config.XSSProtection)
			}

			// Strict-Transport-Security: enforces HTTPS
			if config.HSTSMaxAge > 0 {
				hsts := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
				if config.HSTSIncludeSubdomains {
					hsts += "; includeSubDomains"
				}
				c.SetHeader("Strict-Transport-Security", hsts)
			}

			// Content-Security-Policy: prevents XSS and injection attacks
			if config.ContentSecurityPolicy != "" {
				c.SetHeader("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Referrer-Policy: controls referrer information
			if config.ReferrerPolicy != "" {
				c.SetHeader("Referrer-Policy", config.ReferrerPolicy)
			}

			return next(c)
		}
	}
}
