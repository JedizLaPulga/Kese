package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
)

// CSRFConfig holds configuration for CSRF protection middleware.
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token. Default: 32
	TokenLength int

	// TokenLookup is where to look for the CSRF token.
	// Format: "<source>:<key>"
	// Possible sources: "form", "header"
	// Default: "form:csrf_token"
	TokenLookup string

	// CookieName is the name of the cookie storing the CSRF token.
	// Default: "_csrf"
	CookieName string

	// CookiePath is the path for the CSRF cookie. Default: "/"
	CookiePath string

	// CookieHTTPOnly sets HttpOnly flag on cookie. Default: true
	CookieHTTPOnly bool

	// CookieSameSite sets SameSite attribute. Default: http.SameSiteStrictMode
	CookieSameSite http.SameSite

	// ContextKey is the key to store CSRF token in context. Default: "csrf_token"
	ContextKey string
}

// DefaultCSRFConfig returns the default CSRF configuration.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    32,
		TokenLookup:    "form:csrf_token",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		ContextKey:     "csrf_token",
	}
}

// CSRF returns a middleware that provides CSRF protection.
//
// Example:
//
//	app.Use(middleware.CSRF())
//
//	// In handler, get token for forms:
//	token := c.CSRFToken()
//
//	// In template:
//	<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
func CSRF() kese.MiddlewareFunc {
	return CSRFWithConfig(DefaultCSRFConfig())
}

// CSRFWithConfig returns a CSRF middleware with custom configuration.
func CSRFWithConfig(config CSRFConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Skip CSRF for safe methods
			if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
				// Generate and set token for safe methods
				token, err := generateToken(config.TokenLength)
				if err != nil {
					return err
				}

				// Set cookie
				http.SetCookie(c.Writer, &http.Cookie{
					Name:     config.CookieName,
					Value:    token,
					Path:     config.CookiePath,
					HttpOnly: config.CookieHTTPOnly,
					SameSite: config.CookieSameSite,
				})

				// Store in context for templates
				c.Set(config.ContextKey, token)

				return next(c)
			}

			// For unsafe methods, validate token
			cookieToken, err := c.Cookie(config.CookieName)
			if err != nil || cookieToken == nil {
				return c.Forbidden("CSRF token missing")
			}

			// Extract token from request
			requestToken := extractCSRFToken(c, config.TokenLookup)
			if requestToken == "" {
				return c.Forbidden("CSRF token not provided")
			}

			// Validate tokens match
			if cookieToken.Value != requestToken {
				return c.Forbidden("CSRF token invalid")
			}

			// Store in context
			c.Set(config.ContextKey, cookieToken.Value)

			return next(c)
		}
	}
}

// generateToken generates a random CSRF token.
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// extractCSRFToken extracts CSRF token from request.
func extractCSRFToken(c *context.Context, lookup string) string {
	// Parse lookup format
	if len(lookup) < 6 {
		return ""
	}

	if lookup[:5] == "form:" {
		return c.FormValue(lookup[5:])
	}

	if lookup[:7] == "header:" {
		return c.Header(lookup[7:])
	}

	return ""
}
