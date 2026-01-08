# Kese Tier 2 Features Guide

## Overview

Kese v2.0 includes 11 enterprise-grade features that make it production-ready. All features are **100% backward compatible** and optional.

---

## 🛡️ Security Features

### JWT Authentication

Generate and validate JWT tokens for user authentication.

**Basic Usage**:
```go
import "github.com/JedizLaPulga/kese/auth"

// Generate token
token, err := auth.GenerateToken(auth.Claims{
    "userID": "123",
    "email": "user@example.com",
}, "my-secret-key", 24*time.Hour)

// Protect routes with middleware
app.Use(middleware.JWT("my-secret-key"))

// Access claims in handler
func handler(c *context.Context) error {
    claims := c.Get("jwt_claims").(auth.Claims)
    userID := claims["userID"].(string)
    email := c.Get("email").(string) // Auto-extracted
    
   return c.Success(claims)
}
```

**Advanced**:
```go
// Custom token lookup (cookie, query param)
app.Use(middleware.JWTWith Config(middleware.JWTConfig{
    Secret: "my-secret",
    TokenLookup: "cookie:auth_token",
    SkipFunc: func(c *context.Context) bool {
        return c.Path() =="/login"
    },
}))

// Refresh tokens
newToken, err := auth.RefreshToken(oldToken, secret, 24*time.Hour)
```

---

### CSRF Protection

Prevent cross-site request forgery attacks.

```go
// Enable CSRF
app.Use(middleware.CSRF())

// In handler
func showForm(c *context.Context) error {
    token := c.CSRFToken()
    return c.HTML(200, fmt.Sprintf(`
        <form method="POST">
            <input type="hidden" name="csrf_token" value="%s">
            <button>Submit</button>
        </form>
    `, token))
}

// CSRF validation happens automatically on POST/PUT/DELETE
```

---

### Security Headers

Add security headers automatically.

```go
// Default security headers
app.Use(middleware.SecureHeaders())
// Adds: X-Frame-Options, X-Content-Type-Options, XSS-Protection, HSTS

// Custom configuration
app.Use(middleware.SecureHeadersWithConfig(middleware.SecurityConfig{
    XFrameOptions: "SAMEORIGIN",
    HSTSMaxAge: 63072000, // 2 years
    ContentSecurityPolicy: "default-src 'self'",
}))
```

---

### Rate Limiting

Prevent API abuse with configurable rate limits.

```go
// 100 requests per minute per IP
app.Use(middleware.RateLimit(100, time.Minute))

// Custom rate limiting
app.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
    Limit: 1000,
    Window: time.Hour,
    KeyFunc: func(c *context.Context) string {
        // Rate limit per authenticated user
        if userID := c.Get("userID"); userID != nil {
            return fmt.Sprintf("user:%v", userID)
        }
        return c.Request.RemoteAddr
    },
    SkipFunc: func(c *context.Context) bool {
        // Skip rate limiting for admins
        return c.Get("role") == "admin"
    },
}))

// Headers added automatically:
// X-RateLimit-Limit: 100
// X-RateLimit-Remaining: 95
// Retry-After: 60 (when limit exceeded)
```

---

### Input Sanitization

Protect against XSS and injection attacks.

```go
import "github.com/JedizLaPulga/kese/sanitize"

// In handler
userInput := c.FormValue("bio")
safe := sanitize.HTML(userInput) // Escapes HTML

// Or use context helpers
safe := c.SanitizeHTML(userInput)

// Validators
if !c.IsEmail(email) {
    return c.BadRequest("Invalid email")
}

if !c.IsURL(website) {
    return c.BadRequest("Invalid URL")
}

// Other sanitizers
sanitize.SQL(input)         // Escape SQL
sanitize.Path(input)        // Prevent directory traversal
sanitize.AlphaNumeric(input) // Remove special chars
sanitize.StripTags(html)    // Remove HTML tags
```

---

## ⚡ Performance Features

### Gzip Compression

Automatically compress responses.

```go
// Default compression
app.Use(middleware.Gzip())

// Custom configuration
app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
    Level: 6,  // 1-9, higher = better compression
    MinLength: 512, // Only compress if > 512 bytes
    ExcludedExtensions: []string{".jpg", ".png"}, // Skip these
}))

// Headers set automatically:
// Content-Encoding: gzip
// Vary: Accept-Encoding
```

---

### Response Caching

Cache GET responses in memory.

```go
// Cache all GET requests for 5 minutes
app.Use(middleware.Cache(5 * time.Minute))

// Per-route caching
app.GET("/expensive", expensiveHandler, middleware.Cache(1 * time.Hour))

// Custom cache key
app.Use(middleware.CacheWithConfig(middleware.CacheConfig{
    TTL: 10 * time.Minute,
    KeyFunc: func(c *context.Context) string {
        return c.Path() + ":" + c.Query("lang")
    },
}))

// Cache headers:
// X-Cache: HIT  (served from cache)
// X-Cache: MISS (fresh response, now cached)
```

---

### Graceful Shutdown

Handle shutdown signals properly.

```go
// Instead of app.Run()
app.RunWithShutdown(":8080", 10*time.Second)

// Waits up to 10 seconds for active requests to complete
// Handles SIGINT and SIGTERM signals
```

---

## 📊 Observability Features

### Structured Logging

JSON logging with levels.

```go
// Use the app logger
app.Logger.Info("Server started", "port", 8080, "env", "production")
app.Logger.Error("DB error", "error", err, "query", query)
app.Logger.Debug("Cache hit", "key", cacheKey)
app.Logger.Warn("High memory", "usage", "85%")

// Set log level
app.Logger.SetLevel(logger.DebugLevel)

// Output (JSON):
// {"timestamp":"2026-01-08T06:00:00Z","level":"INFO","message":"Server started","port":8080,"env":"production"}
```

---

### Health Checks

Monitor service health.

```go
// Add health checks
app.AddHealthCheck("database", func() error {
    return db.Ping()
})

app.AddHealthCheck("redis", func() error {
    return redisClient.Ping(ctx).Err()
})

app.AddHealthCheck("api", func() error {
    resp, err := http.Get("https://api.example.com/health")
    if err != nil {
        return err
    }
    if resp.StatusCode != 200 {
        return fmt.Errorf("API returned %d", resp.StatusCode)
    }
    return nil
})

// Expose health endpoint
app.GET("/health", app.HealthHandler())

// Response:
// {"status":"healthy","checks":{"database":"ok","redis":"ok","api":"ok"}}

// Unhealthy = 503 status code
// Healthy = 200 status code
```

---

### Prometheus Metrics

Track request metrics.

```go
// Enable metrics collection
app.Use(middleware.Metrics())

// Expose metrics endpoint
import "github.com/JedizLaPulga/kese/metrics"

app.GET("/metrics", func(c *context.Context) error {
    metrics.Handler().ServeHTTP(c.Writer, c.Request)
    c.SetWritten()
    return nil
})

// Metrics collected:
// - kese_active_requests (gauge)
// - kese_requests_total (counter)
// - kese_errors_total (counter)
// - kese_requests_by_route_total{route="GET /users"} (counter)
// - kese_request_duration_seconds{route="GET /users"} (summary)

// Compatible with Prometheus, Grafana, etc.
```

---

## 🛠️ Developer Experience Features

### Route Groups

Organize routes with prefixes and middleware.

```go
// API v1 with auth
v1 := app.Group("/api/v1", middleware.JWT("secret"))
v1.GET("/users", getUsers)
v1.POST("/users", createUser)

// Admin routes with extra middleware
admin := app.Group("/admin", authMiddleware(), adminMiddleware())
admin.GET("/stats", getStats)
admin.DELETE("/users/:id", deleteUser)

// Nested groups
api := app.Group("/api")
v1 := api.Group("/v1")
v2 := api.Group("/v2")
```

---

### Context Value Storage

Pass data between middleware and handlers.

```go
// In auth middleware
func authMiddleware() kese.MiddlewareFunc {
    return func(next kese.HandlerFunc) kese.HandlerFunc {
        return func(c *context.Context) error {
            user := authenticate(c)
            c.Set("user", user)
            c.Set("role", user.Role)
            return next(c)
        }
    }
}

// In handler
func handler(c *context.Context) error {
    user := c.Get("user").(User)
    role := c.MustGet("role").(string) // Panics if not found
    
    return c.Success(user)
}
```

---

### Response Helpers

Cleaner response code.

```go
// Instead of: c.JSON(200, data)
return c.Success(data)           // 200 OK

// Instead of: c.JSON(201, resource)
return c.Created(resource)       // 201 Created

// Instead of: c.JSON(400, map[string]string{"error": msg})
return c.BadRequest("Invalid input")    // 400
return c.Unauthorized("Login required") // 401
return c.Forbidden("Access denied")     // 403
return c.NotFoundError("Not found")     // 404
return c.InternalError("Server error")  // 500
```

---

### Form & File Uploads

Handle forms and file uploads easily.

```go
// Form data
email := c.FormValue("email")
password := c.PostFormValue("password")
tags := c.FormValues("tags") // Multiple values

// File upload
import "github.com/JedizLaPulga/kese/middleware"

func uploadHandler(c *context.Context) error {
    // Save uploaded file directly
    if err := c.SaveUploadedFile("avatar", "./uploads/avatar.png"); err != nil {
        return c.BadRequest("Upload failed")
    }
    
    return c.Success("File uploaded")
}

// Or handle manually
file, req, err := c.FormFile("avatar")
if err != nil {
    return c.BadRequest("No file")
}
defer file.Close() // file is io.ReadCloser

// Access file info from header
header, _ := req.FormFile("avatar")
filename := header.Filename
size := header.Size
```

---

### Custom Error Handlers

Centralized error handling.

```go
app.SetErrorHandler(func(err error) (int, interface{}) {
    // Handle specific error types
    var validationErr *kese.ValidationError
    if errors.As(err, &validationErr) {
        return 400, map[string]interface{}{
            "error": "validation_failed",
            "fields": validationErr.Errors,
        }
    }
    
    // Handle custom errors
    if err.Error() == "not found" {
        return 404, map[string]string{"error": "Resource not found"}
    }
    
    // Default
    return 500, map[string]string{"error": "Internal server error"}
})

// In handlers, just return errors
return errors.New("not found")  // Becomes 404 automatically
return validationError          // Becomes 400 with fields
```

---

## 🎯 Complete Production Example

```go
package main

import (
    "database/sql"
    "time"
    
    "github.com/JedizLaPulga/kese"
    "github.com/JedizLaPulga/kese/auth"
    "github.com/JedizLaPulga/kese/context"
    "github.com/JedizLaPulga/kese/logger"
    "github.com/JedizLaPulga/kese/metrics"
    "github.com/JedizLaPulga/kese/middleware"
)

func main() {
    app := kese.New()
    
    // Configure logger
    app.Logger.SetLevel(logger.InfoLevel)
    
    // Security middleware
    app.Use(middleware.SecureHeaders())
    app.Use(middleware.CSRF())
    app.Use(middleware.RateLimit(100, time.Minute))
    
    //  Performance middleware
    app.Use(middleware.Gzip())
    app.Use(middleware.Cache(5 * time.Minute))
    
    // Observability middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.Metrics())
    app.Use(middleware.CORS())
    
    // Health checks
    app.AddHealthCheck("database", func() error {
        return db.Ping()
    })
    
    // Public routes
    app.POST("/login", login)
    app.POST("/register", register)
    
    // Health & metrics
    app.GET("/health", app.HealthHandler())
    app.GET("/metrics", func(c *context.Context) error {
        metrics.Handler().ServeHTTP(c.Writer, c.Request)
        c.SetWritten()
        return nil
    })
    
    // Protected API routes
    api := app.Group("/api/v1", middleware.JWT("secret"))
    api.GET("/users", getUsers)
    api.POST("/users", createUser)
    api.GET("/users/:id", getUser)
    api.PUT("/users/:id", updateUser)
    api.DELETE("/users/:id", deleteUser)
    
    // Admin routes
    admin := app.Group("/admin", middleware.JWT("secret"), adminMiddleware())
    admin.GET("/stats", getStats)
    
    // Graceful shutdown
    app.Logger.Info("Starting server", "port", 8080)
    if err := app.RunWithShutdown(":8080", 10*time.Second); err != nil {
        app.Logger.Error("Server error", "error", err)
    }
}
```

---

## 📚 Feature Summary

| Feature | Package | Middleware | Description |
|---------|---------|------------|-------------|
| JWT Auth | `auth` | `middleware.JWT()` | Token-based authentication |
| CSRF | - | `middleware.CSRF()` | Cross-site request forgery protection |
| Security Headers | - | `middleware.SecureHeaders()` | XSS, clickjacking, HSTS |
| Rate Limiting | `ratelimit` | `middleware.RateLimit()` | Request rate limiting |
| Gzip | - | `middleware.Gzip()` | Response compression |
| Caching | `cache` | `middleware.Cache()` | Response caching with TTL |
| Logging | `logger` | - | Structured JSON logging |
| Health Checks | `health` | - | Service health monitoring |
| Metrics | `metrics` | `middleware.Metrics()` | Prometheus-compatible metrics |
| Sanitization | `sanitize` | - | Input sanitization helpers |
| Route Groups | - | - | Route organization |

---

**All features are production-tested and ready to use!** 🚀
