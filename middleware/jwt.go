package middleware

import (
	"strings"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/auth"
	"github.com/JedizLaPulga/kese/context"
)

// JWTConfig holds configuration for JWT middleware.
type JWTConfig struct {
	// Secret is the key used to sign and validate tokens
	Secret string

	// ContextKey is the key used to store claims in context.
	// Default: "jwt_claims"
	ContextKey string

	// TokenLookup is where to look for the token.
	// Format: "<source>:<key>"
	// Possible sources: "header", "query", "cookie"
	// Default: "header:Authorization"
	TokenLookup string

	// SkipFunc allows skipping JWT validation for certain requests.
	// Return true to skip JWT validation for this request.
	SkipFunc func(*context.Context) bool
}

// DefaultJWTConfig returns the default JWT configuration.
func DefaultJWTConfig(secret string) JWTConfig {
	return JWTConfig{
		Secret:      secret,
		ContextKey:  "jwt_claims",
		TokenLookup: "header:Authorization",
		SkipFunc:    nil,
	}
}

// JWT returns a middleware that validates JWT tokens.
//
// Example:
//
//	app.Use(middleware.JWT("my-secret-key"))
//
//	// In handler
//	claims := c.Get("jwt_claims").(auth.Claims)
//	userID := claims["userID"].(string)
func JWT(secret string) kese.MiddlewareFunc {
	return JWTWithConfig(DefaultJWTConfig(secret))
}

// JWTWithConfig returns a JWT middleware with custom configuration.
//
// Example:
//
//	app.Use(middleware.JWTWithConfig(JWTConfig{
//	    Secret: "my-secret",
//	    TokenLookup: "cookie:token",
//	    SkipFunc: func(c *context.Context) bool {
//	        // Skip JWT for public routes
//	        return c.Path() == "/login" || c.Path() == "/register"
//	    },
//	}))
func JWTWithConfig(config JWTConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Check if we should skip JWT validation
			if config.SkipFunc != nil && config.SkipFunc(c) {
				return next(c)
			}

			// Extract token from request
			token, err := extractToken(c, config.TokenLookup)
			if err != nil {
				return c.Unauthorized("missing or invalid token")
			}

			// Validate token
			claims, err := auth.ValidateToken(token, config.Secret)
			if err != nil {
				if err == auth.ErrTokenExpired {
					return c.Unauthorized("token has expired")
				}
				return c.Unauthorized("invalid token")
			}

			// Store claims in context
			c.Set(config.ContextKey, claims)

			// Optional: Store individual claims for convenience
			if userID, ok := claims["userID"]; ok {
				c.Set("userID", userID)
			}
			if email, ok := claims["email"]; ok {
				c.Set("email", email)
			}

			return next(c)
		}
	}
}

// extractToken extracts JWT token from request based on TokenLookup config.
func extractToken(c *context.Context, lookup string) (string, error) {
	parts := strings.Split(lookup, ":")
	if len(parts) != 2 {
		return "", nil
	}

	source := parts[0]
	key := parts[1]

	switch source {
	case "header":
		authHeader := c.Header(key)
		return auth.ExtractTokenFromHeader(authHeader)

	case "query":
		token := c.Query(key)
		if token == "" {
			return "", nil
		}
		return token, nil

	case "cookie":
		cookie, err := c.Cookie(key)
		if err != nil {
			return "", err
		}
		return cookie.Value, nil
	}

	return "", nil
}
