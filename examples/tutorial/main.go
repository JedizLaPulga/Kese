package main

import (
	"strconv"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/middleware"
)

// Global store
var store = NewTodoStore()

func main() {
	app := kese.New()

	// Middleware
	app.Use(middleware.Logger(app.Logger))
	app.Use(middleware.Recovery(app.Logger))
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
			"GET /todos":        "List all todos",
			"GET /todos/:id":    "Get a specific todo",
			"POST /todos":       "Create a new todo",
			"PUT /todos/:id":    "Update a todo",
			"DELETE /todos/:id": "Delete a todo",
		},
	})
}

// GET /todos - List all todos
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
