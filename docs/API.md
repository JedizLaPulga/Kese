# Kese API Documentation

Welcome to Kese! This guide provides comprehensive API documentation for building web applications with the Kese framework.

## Table of Contents
- [Getting Started](#getting-started)
- [Core API](#core-api)
- [Context](#context)
- [Routing](#routing)
- [Middleware](#middleware)
- [Examples](#examples)

---

## Getting Started

### Installation

```bash
go get github.com/JedizLaPulga/kese
```

### Quick Example

```go
package main

import (
    "github.com/JedizLaPulga/kese"
    "github.com/JedizLaPulga/kese/context"
)

func main() {
    app := kese.New()
    
    app.GET("/", func(c *context.Context) error {
        return c.JSON(200, map[string]string{"message": "Hello, Kese!"})
    })
    
    app.Run(":8080")
}
```

---

## Core API

### Creating an App

```go
app := kese.New()
```

Creates a new Kese application instance.

### HTTP Methods

Register routes for specific HTTP methods:

```go
app.GET(path string, handler HandlerFunc)
app.POST(path string, handler HandlerFunc)
app.PUT(path string, handler HandlerFunc)
app.DELETE(path string, handler HandlerFunc)
app.PATCH(path string, handler HandlerFunc)
app.OPTIONS(path string, handler HandlerFunc)
app.HEAD(path string, handler HandlerFunc)
```

### Middleware

```go
app.Use(middleware ...MiddlewareFunc)
```

Add middleware to your application. Middleware is executed in the order it is registered.

### Running the Server

```go
// HTTP
err := app.Run(":8080")

// HTTPS
err := app.RunTLS(":443", "cert.pem", "key.pem")
```

---

## Context

The context provides methods for handling requests and sending responses.

### Request Methods

#### URL Parameters

```go
// Route: /users/:id
id := c.Param("id")
```

#### Query Strings

```go
// URL: /search?q=golang
query := c.Query("q")

// With default value
query := c.QueryDefault("q", "default")
```

#### Headers

```go
auth := c.Header("Authorization")
contentType := c.Header("Content-Type")
```

#### Request Body

```go
var user User
if err := c.Body(&user); err != nil {
    return c.JSON(400, map[string]string{"error": "Invalid body"})
}
```

#### Raw Body

```go
data, err := c.BodyBytes()
```

#### Cookies

```go
cookie, err := c.Cookie("session")
```

#### HTTP Method & Path

```go
method := c.Method()  // "GET", "POST", etc.
path := c.Path()      // "/users/123"
```

### Response Methods

#### JSON

```go
// Standard JSON
c.JSON(200, map[string]interface{}{
    "message": "Success",
    "data": data,
})

// Pretty-printed JSON (for debugging)
c.JSONPretty(200, data)
```

#### Plain Text

```go
c.String(200, "Hello, World!")
```

#### HTML

```go
c.HTML(200, "<h1>Hello, World!</h1>")
```

#### Raw Bytes

```go
c.Bytes(200, "application/pdf", pdfData)
```

#### No Content

```go
c.NoContent()  // Returns 204 No Content
```

#### Redirects

```go
c.Redirect(301, "/new-location")  // Permanent redirect
c.Redirect(302, "/temporary")     // Temporary redirect
```

#### Headers

```go
c.SetHeader("X-Custom-Header", "value")
c.SetHeader("Content-Type", "application/json")
```

#### Cookies

```go
cookie := &http.Cookie{
    Name:  "session",
    Value: "token123",
    Path:  "/",
}
c.SetCookie(cookie)
```

#### Status Code

```go
c.Status(201)  // Set status code
code := c.StatusCode()  // Get status code
```

### Direct Access (Advanced)

For advanced use cases, access the underlying primitives:

```go
c.Request  // *http.Request
c.Writer   // http.ResponseWriter
```

---

## Routing

### Static Routes

```go
app.GET("/users", handleUsers)
app.GET("/about", handleAbout)
```

### Dynamic Parameters

```go
// Single parameter
app.GET("/users/:id", handleUser)

// Multiple parameters
app.GET("/users/:userId/posts/:postId", handlePost)

// Mixed
app.GET("/api/v1/users/:id/profile", handleProfile)
```

### Route Handlers

Handler signature:

```go
func(c *context.Context) error
```

Example:

```go
func handleUser(c *context.Context) error {
    id := c.Param("id")
    
    user, err := getUser(id)
    if err != nil {
        return c.JSON(404, map[string]string{"error": "User not found"})
    }
    
    return c.JSON(200, user)
}
```

---

## Middleware

### Using Middleware

```go
app.Use(middleware.Logger())
app.Use(middleware.Recovery())
app.Use(middleware.CORS())
```

### Built-in Middleware

#### Logger

Logs HTTP requests with method, path, status, and duration:

```go
app.Use(middleware.Logger())
```

#### Recovery

Recovers from panics and returns 500 errors:

```go
app.Use(middleware.Recovery())
```

#### CORS

Adds CORS headers for cross-origin requests:

```go
// Default CORS (allows all origins)
app.Use(middleware.CORS())

// Custom CORS
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{"GET", "POST"},
    AllowHeaders: []string{"Authorization"},
}))
```

#### Request ID

Adds unique request ID to each request:

```go
app.Use(middleware.RequestID())
```

### Custom Middleware

Create your own middleware:

```go
func Auth() kese.MiddlewareFunc {
    return func(next kese.HandlerFunc) kese.HandlerFunc {
        return func(c *context.Context) error {
            token := c.Header("Authorization")
            if token == "" {
                return c.JSON(401, map[string]string{
                    "error": "Unauthorized",
                })
            }
            
            // Validate token...
            
            return next(c)
        }
    }
}

// Use it
app.Use(Auth())
```

---

## Examples

### REST API

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    app := kese.New()
    
    app.GET("/users", listUsers)
    app.GET("/users/:id", getUser)
    app.POST("/users", createUser)
    app.PUT("/users/:id", updateUser)
    app.DELETE("/users/:id", deleteUser)
    
    app.Run(":8080")
}

func listUsers(c *context.Context) error {
    users := []User{/* fetch from DB */}
    return c.JSON(200, users)
}

func getUser(c *context.Context) error {
    id := c.Param("id")
    user := User{/* fetch from DB */}
    return c.JSON(200, user)
}

func createUser(c *context.Context) error {
    var user User
    if err := c.Body(&user); err != nil {
        return c.JSON(400, map[string]string{"error": "Invalid body"})
    }
    
    // Save to DB...
    
    return c.JSON(201, user)
}
```

### With Middleware

```go
func main() {
    app := kese.New()
    
    // Global middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    
    // Protected routes
    app.Use(Auth())
    app.GET("/profile", getProfile)
    
    app.Run(":8080")
}
```

### Error Handling

```go
func handleUser(c *context.Context) error {
    id := c.Param("id")
    
    user, err := db.GetUser(id)
    if err != nil {
        if err == ErrNotFound {
            return c.JSON(404, map[string]string{
                "error": "User not found",
            })
        }
        // Framework will catch this and return 500
        return fmt.Errorf("database error: %w", err)
    }
    
    return c.JSON(200, user)
}
```

---

## Best Practices

1. **Return errors from handlers** - Let the framework handle internal errors
2. **Use middleware for cross-cutting concerns** - Logging, auth, CORS, etc.
3. **Validate input early** - Check params and body before processing
4. **Use proper HTTP status codes** - 200, 201, 400, 401, 404, 500, etc.
5. **Structure your routes** - Group related endpoints
6. **Access underlying primitives when needed** - `c.Request` and `c.Writer` are there for you

---

## Contributing

See [PROJECT_RULES.md](../PROJECT_RULES.md) for development guidelines.

## License

MIT License
