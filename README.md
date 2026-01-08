# Kese

> A modern, fast, and **enterprise-ready** Go web framework inspired by FastAPI

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Test Coverage](https://img.shields.io/badge/coverage-87%25-brightgreen.svg)](https://github.com/JedizLaPulga/kese)
[![Tests](https://img.shields.io/badge/tests-67%20passing-brightgreen.svg)](https://github.com/JedizLaPulga/kese)
[![Completion](https://img.shields.io/badge/completion-90%25-blue.svg)](https://github.com/JedizLaPulga/kese)

## 🎯 Overview

Kese is a **production-ready**, high-performance web framework for Go that brings the elegant developer experience of FastAPI to the Go ecosystem. Built entirely on Go's standard library with **enterprise-grade features** including JWT auth, rate limiting, caching, metrics, and more.

### Philosophy

**Do much, but stay out of the way.**

Kese gives you enterprise features without the complexity. Zero dependencies, clean API, and backward compatible updates.

## ✨ Features

### 🚀 Core Features
- **Lightning Fast**: Radix tree routing with O(log n) lookup
- **Zero Dependencies**: Pure Go standard library
- **Type-Safe**: Leverage Go's type system
- **Fully Tested**: 67 tests with 87%+ coverage
- **Modular Design**: Use only what you need
- **Beautiful Landing Page**: Professional welcome screen

### 🔐 Security & Auth
- **JWT Authentication**: Token generation, validation, refresh
- **CSRF Protection**: Cookie-based CSRF defense
- **Security Headers**: Clickjacking, HSTS protection (X-XSS-Protection removed as deprecated)
- **Rate Limiting**: Prevent API abuse with configurable limits
- **Input Sanitization**: HTML, path sanitization helpers

### ⚡ Performance
- **Gzip Compression**: Automatic response compression
- **Response Caching**: In-memory caching with TTL
- **Graceful Shutdown**: Zero-downtime deployments

### 📊 Observability  
- **Structured Logging**: JSON logs with levels
- **Prometheus Metrics**: Request counts, duration, errors
- **Health Checks**: Liveness & readiness endpoints

### 🛠️ Developer Experience
- **Route Groups**: Organize routes with prefixes
- **Context Value Storage**: Pass data between middleware
- **Response Helpers**: Success(), BadRequest(), etc.
- **Form & File Uploads**: Easy file handling
- **Template Rendering**: Server-side HTML
- **Custom Error Handlers**: Centralized error handling

## 🚀 Quick Start

### Installation

```bash
go get github.com/JedizLaPulga/kese
```

### Hello World

```go
package main

import (
    "github.com/JedizLaPulga/kese"
    "github.com/JedizLaPulga/kese/context"
    "github.com/JedizLaPulga/kese/middleware"
)

func main() {
    app := kese.New()
    
    // Add middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.CORS())
    
    // Define routes
    app.GET("/", func(c *context.Context) error {
        return c.Success(map[string]string{
            "message": "Hello from Kese!",
            "version": "v1.0",
        })
    })
    
    // Run with graceful shutdown
    app.RunWithShutdown(":8080", 10*time.Second)
}
```

### Production Example

```go
package main

import (
    "time"
    "github.com/JedizLaPulga/kese"
    "github.com/JedizLaPulga/kese/context"
    "github.com/JedizLaPulga/kese/middleware"
)

func main() {
    app := kese.New()
    
    // Production middleware stack
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.SecureHeaders())
    app.Use(middleware.Gzip())
    app.Use(middleware.RateLimit(100, time.Minute))  // 100 req/min
    app.Use(middleware.Metrics())
    
    // API routes with JWT auth
    api := app.Group("/api/v1", middleware.JWT("your-secret-key"))
    api.GET("/users", getUsers)
    api.POST("/users", createUser)
    
    // Health & metrics endpoints
    app.GET("/health", app.HealthHandler())
    app.GET("/metrics", metricsHandler)
    
    // Graceful shutdown
    app.RunWithShutdown(":8080", 10*time.Second)
}
```

## 📚 Documentation

### Routing

```go
// Basic routes
app.GET("/users", getUsers)
app.POST("/users", createUser)
app.PUT("/users/:id", updateUser)
app.DELETE("/users/:id", deleteUser)

// Route groups with middleware
api := app.Group("/api/v1", authMiddleware())
admin := app.Group("/admin", authMiddleware(), adminMiddleware())

// Static files
app.StaticFile("/", "./templates/index.html")
app.Static("/assets", "./public")
```

### Context API

```go
func handler(c *context.Context) error {
    // Request data
    id := c.Param("id")
    email := c.Query("email")
    token := c.Header("Authorization")
    
    // Parse body
    var user User
    if err := c.Body(&user); err != nil {
        return c.BadRequest("Invalid JSON")
    }
    
    // Validate
    if !c.IsEmail(user.Email) {
        return c.BadRequest("Invalid email")
    }
    
    // Sanitize
    user.Bio = c.SanitizeHTML(user.Bio)
    
    // Store in context
    c.Set("user", user)
    
    // Response helpers
    return c.Success(user)           // 200 OK
    return c.Created(user)           // 201 Created
    return c.BadRequest("message")   // 400
    return c.Unauthorized("message") // 401
    return c.NotFoundError("message")// 404
}
```

### Middleware

```go
// Security
app.Use(middleware.SecureHeaders())
app.Use(middleware.CSRF())
app.Use(middleware.JWT("secret"))
app.Use(middleware.RateLimit(100, time.Minute))

// Performance
app.Use(middleware.Gzip())
app.Use(middleware.Cache(5 * time.Minute))

// Observability
app.Use(middleware.Logger())
app.Use(middleware.Metrics())
app.Use(middleware.Recovery())

// CORS
app.Use(middleware.CORS())
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{"GET", "POST"},
}))
```

### Authentication (JWT)

```go
import "github.com/JedizLaPulga/kese/auth"

// Generate token
token, err := auth.GenerateToken(auth.Claims{
    "userID": "123",
    "email": "user@example.com",
}, "secret-key", 24*time.Hour)

// Protect routes
app.Use(middleware.JWT("secret-key"))

// In handler
claims := c.Get("jwt_claims").(auth.Claims)
userID := claims["userID"].(string)
```

### Rate Limiting

```go
// Simple rate limit
app.Use(middleware.RateLimit(100, time.Minute)) // 100 req/min

// Custom configuration
app.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
    Limit: 1000,
    Window: time.Hour,
    KeyFunc: func(c *context.Context) string {
        // Rate limit per user instead of IP
        return c.Get("userID").(string)
    },
}))
```

### Health Checks

```go
// Add health checks
app.AddHealthCheck("database", func() error {
    return db.Ping()
})

app.AddHealthCheck("redis", func() error {
    return redis.Ping()
})

// Expose endpoint
app.GET("/health", app.HealthHandler())
// Returns: {"status":"healthy","checks":{"database":"ok","redis":"ok"}}
```

### Metrics & Monitoring

```go
// Enable metrics collection
app.Use(middleware.Metrics())

// Expose Prometheus endpoint
app.GET("/metrics", metricsHandler)

// Metrics tracked:
// - kese_active_requests
// - kese_requests_total
// - kese_errors_total  
// - kese_requests_by_route_total
// - kese_request_duration_seconds
```

### Response Caching

```go
// Cache all GET requests for 5 minutes
app.Use(middleware.Cache(5 * time.Minute))

// Per-route caching
app.GET("/expensive", expensiveHandler, middleware.Cache(1 * time.Hour))

// Cache headers added automatically:
// X-Cache: HIT or MISS
```

### Logging

```go
// Use structured logger
app.Logger.Info("Server started", "port", 8080, "env", "production")
app.Logger.Error("Database error", "error", err, "query", query)

// Configure log level
app.Logger.SetLevel(logger.DebugLevel)
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Stress test
go test -count=1000 ./...

# Coverage: 87%+
# Tests passing: 67/67
```

## 📊 Test Coverage

| Package | Coverage | Tests |
|---------|----------|-------|
| kese | 83% | 20 |
| context | 87% | 26 |
| middleware | 100% | 10 |
| router | 97% | 11 |

## 🏗️ Project Structure

```
kese/
├── auth/           # JWT authentication
├── cache/          # Response caching
├── context/        # Request/response context
├── health/         # Health check system
├── logger/         # Structured logging
├── metrics/        # Prometheus metrics
├── middleware/     # Built-in middleware
├── ratelimit/      # Rate limiting store
├── router/         # Radix tree router
├── sanitize/       # Input sanitization
├── examples/       # Example applications
├── docs/           # Documentation
└── templates/      # HTML templates
```

## 🚦 Roadmap

- [x] **v1.0** - Core framework (Routing, Context, Middleware)
- [x] **v1.5** - Edge case fixes & stability
- [x] **v2.0** - Enterprise features (Tier 2 complete!)
  - [x] JWT Authentication
  - [x] Rate Limiting
  - [x] Security Headers & CSRF
  - [x] Metrics & Health Checks
  - [x] Caching & Compression
  - [x] Structured Logging
- [ ] **v2.5** - Advanced features (Optional)
  - [ ] WebSocket support
  - [ ] GraphQL support
  - [ ] OpenAPI/Swagger generation

## 🎯 Development Principles

1. **Standard Library Only**: No external dependencies
2. **Test Everything**: Maintain 80%+ test coverage
3. **Stay Modular**: Features are independent and optional
4. **Document Well**: Code is self-documenting with examples
5. **Zero Breaking Changes**: Backward compatibility guaranteed

## 📖 Examples

Check out the [examples](./examples) directory for complete applications:
- **Basic**: Simple API with middleware
- **Tutorial**: Full Todo API with CRUD operations

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🌟 Star History

If you find Kese useful, please consider giving it a star! ⭐

---

**Built with ❤️ using only Go's standard library**

**Status**: Production-ready • 90% feature complete • Enterprise-grade
