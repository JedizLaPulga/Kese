# Kese Framework Tutorial: Build a Todo API

Welcome! This hands-on tutorial will teach you how to build a complete REST API using the Kese framework. By the end, you'll have built a fully functional Todo application with all CRUD operations.

## What You'll Learn

- ✅ Setting up a Kese project
- ✅ Creating routes and handlers
- ✅ Using middleware (Logger, Recovery, CORS)
- ✅ Working with JSON requests and responses
- ✅ Implementing CRUD operations
- ✅ Error handling
- ✅ Testing your API

## Prerequisites

- Go 1.21 or higher installed
- Basic understanding of Go
- A text editor or IDE
- A tool to test APIs (curl, Postman, or your browser)

---

## Part 1: Getting Started

### Step 1: Create Your Project

```bash
mkdir kese-todo-tutorial
cd kese-todo-tutorial
go mod init github.com/yourusername/kese-todo-tutorial
```

### Step 2: Install Kese

```bash
go get github.com/JedizLaPulga/kese
```

### Step 3: Create Your First Server

Create a file called `main.go`:

```go
package main

import (
    "github.com/JedizLaPulga/kese"
    "github.com/JedizLaPulga/kese/context"
)

func main() {
    // Create a new Kese app
    app := kese.New()
    
    // Define a simple route
    app.GET("/", func(c *context.Context) error {
        return c.JSON(200, map[string]string{
            "message": "Welcome to Kese Todo API!",
        })
    })
    
    // Start the server
    app.Run(":8080")
}
```

### Step 4: Run Your Server

```bash
go run main.go
```

Visit `http://localhost:8080` in your browser. You should see the beautiful Kese landing page!

Visit `http://localhost:8080/api/health` (or any API route if you set one up in examples) or test with curl:

```bash
curl http://localhost:8080/
```

You should see:
```json
{"message": "Welcome to Kese Todo API!"}
```

🎉 **Congratulations!** You've created your first Kese server!

---

## Part 2: Adding Middleware

Middleware runs before your route handlers. Let's add logging, panic recovery, and CORS support.

Update your `main.go`:

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
    app.Use(middleware.Logger())      // Logs all requests
    app.Use(middleware.Recovery())    // Recovers from panics
    app.Use(middleware.CORS())        // Enables CORS
    app.Use(middleware.RequestID())   // Adds unique request IDs
    
    app.GET("/", func(c *context.Context) error {
        return c.JSON(200, map[string]string{
            "message": "Welcome to Kese Todo API!",
        })
    })
    
    app.Run(":8080")
}
```

**What each middleware does:**
- **Logger**: Prints request method, path, status, and duration
- **Recovery**: Catches panics and returns 500 errors instead of crashing
- **CORS**: Allows cross-origin requests (useful for frontend apps)
- **RequestID**: Adds a unique ID to each request for tracking

Restart your server and make a request. You'll now see logs in the console!

---

## Part 3: Building the Todo Data Model

Let's create a Todo struct and an in-memory database.

Create a new file `todo.go`:

```go
package main

import (
    "sync"
    "time"
)

// Todo represents a todo item
type Todo struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Completed bool      `json:"completed"`
    CreatedAt time.Time `json:"created_at"`
}

// TodoStore is our in-memory database
type TodoStore struct {
    mu    sync.RWMutex
    todos map[int]*Todo
    nextID int
}

// NewTodoStore creates a new todo store
func NewTodoStore() *TodoStore {
    return &TodoStore{
        todos: make(map[int]*Todo),
        nextID: 1,
    }
}

// Create adds a new todo
func (s *TodoStore) Create(title string) *Todo {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    todo := &Todo{
        ID:        s.nextID,
        Title:     title,
        Completed: false,
        CreatedAt: time.Now(),
    }
    
    s.todos[s.nextID] = todo
    s.nextID++
    
    return todo
}

// GetAll returns all todos
func (s *TodoStore) GetAll() []*Todo {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    todos := make([]*Todo, 0, len(s.todos))
    for _, todo := range s.todos {
        todos = append(todos, todo)
    }
    
    return todos
}

// Get returns a todo by ID
func (s *TodoStore) Get(id int) (*Todo, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    todo, exists := s.todos[id]
    return todo, exists
}

// Update updates a todo
func (s *TodoStore) Update(id int, title string, completed bool) (*Todo, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    todo, exists := s.todos[id]
    if !exists {
        return nil, false
    }
    
    todo.Title = title
    todo.Completed = completed
    
    return todo, true
}

// Delete removes a todo
func (s *TodoStore) Delete(id int) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    _, exists := s.todos[id]
    if !exists {
        return false
    }
    
    delete(s.todos, id)
    return true
}
```

**What we've created:**
- A `Todo` struct with JSON tags
- A thread-safe in-memory store using `sync.RWMutex`
- CRUD methods: Create, GetAll, Get, Update, Delete

---

## Part 4: Implementing CRUD Routes

Now let's create handlers for our Todo API. Update `main.go`:

```go
package main

import (
    "strconv"
    
    "github.com/JedizLaPulga/kese"
    "github.com/JedizLaPulga/kese/context"
    "github.com/JedizLaPulga/kese/middleware"
)

// Global store (in production, use dependency injection)
var store = NewTodoStore()

func main() {
    app := kese.New()
    
    // Middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.CORS())
    app.Use(middleware.RequestID())
    
    // Routes
    app.GET("/", handleHome)
    app.GET("/todos", handleGetTodos)
    app.GET("/todos/:id", handleGetTodo)
    app.POST("/todos", handleCreateTodo)
    app.PUT("/todos/:id", handleUpdateTodo)
    app.DELETE("/todos/:id", handleDeleteTodo)
    
    app.Run(":8080")
}

// Home handler
func handleHome(c *context.Context) error {
    return c.JSON(200, map[string]interface{}{
        "message": "Welcome to Kese Todo API!",
        "version": "1.0.0",
        "endpoints": map[string]string{
            "GET /todos":       "List all todos",
            "GET /todos/:id":   "Get a specific todo",
            "POST /todos":      "Create a new todo",
            "PUT /todos/:id":   "Update a todo",
            "DELETE /todos/:id": "Delete a todo",
        },
    })
}

// GET /todos - List all todos
func handleGetTodos(c *context.Context) error {
    todos := store.GetAll()
    return c.JSON(200, map[string]interface{}{
        "todos": todos,
        "count": len(todos),
    })
}

// GET /todos/:id - Get a specific todo
func handleGetTodo(c *context.Context) error {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return c.JSON(400, map[string]string{
            "error": "Invalid todo ID",
        })
    }
    
    todo, exists := store.Get(id)
    if !exists {
        return c.JSON(404, map[string]string{
            "error": "Todo not found",
        })
    }
    
    return c.JSON(200, todo)
}

// POST /todos - Create a new todo
func handleCreateTodo(c *context.Context) error {
    var input struct {
        Title string `json:"title"`
    }
    
    if err := c.Body(&input); err != nil {
        return c.JSON(400, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    if input.Title == "" {
        return c.JSON(400, map[string]string{
            "error": "Title is required",
        })
    }
    
    todo := store.Create(input.Title)
    
    return c.JSON(201, map[string]interface{}{
        "message": "Todo created successfully",
        "todo":    todo,
    })
}

// PUT /todos/:id - Update a todo
func handleUpdateTodo(c *context.Context) error {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return c.JSON(400, map[string]string{
            "error": "Invalid todo ID",
        })
    }
    
    var input struct {
        Title     string `json:"title"`
        Completed bool   `json:"completed"`
    }
    
    if err := c.Body(&input); err != nil {
        return c.JSON(400, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    if input.Title == "" {
        return c.JSON(400, map[string]string{
            "error": "Title is required",
        })
    }
    
    todo, exists := store.Update(id, input.Title, input.Completed)
    if !exists {
        return c.JSON(404, map[string]string{
            "error": "Todo not found",
        })
    }
    
    return c.JSON(200, map[string]interface{}{
        "message": "Todo updated successfully",
        "todo":    todo,
    })
}

// DELETE /todos/:id - Delete a todo
func handleDeleteTodo(c *context.Context) error {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return c.JSON(400, map[string]string{
            "error": "Invalid todo ID",
        })
    }
    
    if !store.Delete(id) {
        return c.JSON(404, map[string]string{
            "error": "Todo not found",
        })
    }
    
    return c.JSON(200, map[string]string{
        "message": "Todo deleted successfully",
    })
}
```

---

## Part 5: Testing Your API

### Start the server:
```bash
go run main.go todo.go
```

### Test with curl:

**1. Get all todos (empty initially):**
```bash
curl http://localhost:8080/todos
```
Response:
```json
{"count":0,"todos":[]}
```

**2. Create a todo:**
```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Kese Framework"}'
```
Response:
```json
{
  "message": "Todo created successfully",
  "todo": {
    "id": 1,
    "title": "Learn Kese Framework",
    "completed": false,
    "created_at": "2026-01-07T22:00:00Z"
  }
}
```

**3. Get all todos:**
```bash
curl http://localhost:8080/todos
```
Response:
```json
{
  "count": 1,
  "todos": [
    {
      "id": 1,
      "title": "Learn Kese Framework",
      "completed": false,
      "created_at": "2026-01-07T22:00:00Z"
    }
  ]
}
```

**4. Get a specific todo:**
```bash
curl http://localhost:8080/todos/1
```

**5. Update a todo:**
```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Master Kese Framework","completed":true}'
```

**6. Delete a todo:**
```bash
curl -X DELETE http://localhost:8080/todos/1
```

---

## Part 6: Key Concepts Explained

### 1. **Route Parameters**
```go
app.GET("/todos/:id", handleGetTodo)
```
Access with: `c.Param("id")`

### 2. **Request Body Parsing**
```go
var input struct {
    Title string `json:"title"`
}
c.Body(&input)  // Automatically parses JSON
```

### 3. **JSON Responses**
```go
return c.JSON(200, data)  // Status code + data
```

### 4. **Error Handling**
```go
if err != nil {
    return c.JSON(400, map[string]string{"error": "Bad request"})
}
```

### 5. **Middleware**
Executes in order:
```
Request → Logger → Recovery → CORS → RequestID → Your Handler → Response
```

---

## Part 7: Adding More Features

### Feature 1: Query Parameters (Filtering)

```go
// GET /todos?completed=true
func handleGetTodos(c *context.Context) error {
    todos := store.GetAll()
    
    // Filter by completed status if provided
    if completedStr := c.Query("completed"); completedStr != "" {
        completed := completedStr == "true"
        filtered := []*Todo{}
        for _, todo := range todos {
            if todo.Completed == completed {
                filtered = append(filtered, todo)
            }
        }
        todos = filtered
    }
    
    return c.JSON(200, map[string]interface{}{
        "todos": todos,
        "count": len(todos),
    })
}
```

Test it:
```bash
curl "http://localhost:8080/todos?completed=true"
```

### Feature 2: Custom Headers

```go
func handleGetTodos(c *context.Context) error {
    todos := store.GetAll()
    
    // Add custom header
    c.SetHeader("X-Total-Count", strconv.Itoa(len(todos)))
    
    return c.JSON(200, map[string]interface{}{
        "todos": todos,
    })
}
```

### Feature 3: Health Check Endpoint

```go
app.GET("/health", func(c *context.Context) error {
    return c.JSON(200, map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
        "version": "1.0.0",
    })
})
```

---

## Part 8: Project Structure (Best Practices)

For larger projects, organize your code like this:

```
kese-todo-tutorial/
├── main.go                 # Entry point
├── handlers/
│   ├── todos.go           # Todo handlers
│   └── health.go          # Health check
├── models/
│   └── todo.go            # Todo model
├── store/
│   └── memory.go          # In-memory store
└── middleware/
    └── auth.go            # Custom middleware
```

---

## Part 9: Production Considerations

### 1. **Add Graceful Shutdown**
```go
// You can use http.Server for more control
server := &http.Server{
    Addr:    ":8080",
    Handler: app,
}
server.ListenAndServe()
```

### 2. **Environment Variables**
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
app.Run(":" + port)
```

### 3. **Add Authentication**
Create a custom middleware:
```go
func authMiddleware() kese.MiddlewareFunc {
    return func(next kese.HandlerFunc) kese.HandlerFunc {
        return func(c *context.Context) error {
            token := c.Header("Authorization")
            if token != "Bearer valid-token" {
                return c.JSON(401, map[string]string{
                    "error": "Unauthorized",
                })
            }
            return next(c)
        }
    }
}

app.Use(authMiddleware())
```

---

## Part 10: Next Steps

Now that you've built a Todo API, try these challenges:

### Challenge 1: Add Pagination
```go
// GET /todos?page=1&limit=10
```

### Challenge 2: Add Sorting
```go
// GET /todos?sort=created_at&order=desc
```

### Challenge 3: Add Search
```go
// GET /todos?search=kese
```

### Challenge 4: Add Tags
```go
type Todo struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Completed bool      `json:"completed"`
    Tags      []string  `json:"tags"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Challenge 5: Connect to a Real Database
- PostgreSQL with `database/sql`
- MongoDB with `mongo-go-driver`
- SQLite for simplicity

---

## Complete Code Reference

### Directory Structure:
```
kese-todo-tutorial/
├── main.go
└── todo.go
```

### main.go (Complete):
[See Part 4 code above]

### todo.go (Complete):
[See Part 3 code above]

---

## Troubleshooting

### Issue: "Port already in use"
**Solution**: Change the port or kill the process:
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <pid> /F

# macOS/Linux
lsof -i :8080
kill -9 <pid>
```

### Issue: "Module not found"
**Solution**: Run `go mod tidy`

### Issue: JSON not parsing
**Solution**: Check Content-Type header:
```bash
curl -H "Content-Type: application/json" ...
```

---

## Additional Resources

- **Full API Documentation**: [docs/API.md](API.md)
- **Middleware Documentation**: See middleware package
- **Examples**: [examples/basic/](../examples/basic/)

---

## Summary

🎉 **Congratulations!** You've built a complete REST API with Kese!

**What you learned:**
- ✅ Setting up a Kese project
- ✅ Using middleware
- ✅ Creating RESTful routes
- ✅ Handling JSON requests/responses
- ✅ Implementing CRUD operations
- ✅ Error handling
- ✅ Testing APIs

**Your API supports:**
- Creating todos
- Reading todos (all or by ID)
- Updating todos
- Deleting todos
- Filtering and querying

**Next steps:**
- Build more complex apps
- Add database integration
- Deploy to production
- Build a frontend with React/Vue

---

**Happy coding with Kese! 🚀**

*For questions or issues, visit: https://github.com/JedizLaPulga/kese*
