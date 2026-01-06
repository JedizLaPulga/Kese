package main

import (
	"log"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/middleware"
)

// User represents a simple user model for demonstration
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// In-memory "database" for demonstration
var users = map[int]User{
	1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
	2: {ID: 2, Name: "Bob", Email: "bob@example.com"},
	3: {ID: 3, Name: "Charlie", Email: "charlie@example.com"},
}

func main() {
	// Create a new Kese app
	app := kese.New()

	// Add global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestID())

	// Define routes
	app.GET("/", handleHome)
	app.GET("/health", handleHealth)

	// User routes
	app.GET("/users", handleGetUsers)
	app.GET("/users/:id", handleGetUser)
	app.POST("/users", handleCreateUser)

	// Example of error handling
	app.GET("/panic", handlePanic)

	// Start server
	log.Println("Starting Kese example server...")
	if err := app.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// handleHome is the root endpoint
func handleHome(c *context.Context) error {
	return c.JSON(200, map[string]interface{}{
		"message": "Welcome to Kese Framework! 🚀",
		"version": "0.1.0",
		"endpoints": map[string]string{
			"health":      "/health",
			"users":       "/users",
			"user by id":  "/users/:id",
			"create user": "POST /users",
		},
	})
}

// handleHealth is a health check endpoint
func handleHealth(c *context.Context) error {
	return c.JSON(200, map[string]interface{}{
		"status": "healthy",
	})
}

// handleGetUsers returns all users
func handleGetUsers(c *context.Context) error {
	// Convert map to slice
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}

	return c.JSON(200, map[string]interface{}{
		"users": userList,
		"count": len(userList),
	})
}

// handleGetUser returns a single user by ID
func handleGetUser(c *context.Context) error {
	id := c.Param("id")

	// In a real app, you'd parse the ID and fetch from a database
	// For this example, we'll just return a simple response
	return c.JSON(200, map[string]interface{}{
		"message": "Fetching user with ID: " + id,
		"note":    "In production, this would fetch from a database",
	})
}

// handleCreateUser creates a new user
func handleCreateUser(c *context.Context) error {
	var newUser User

	// Parse request body
	if err := c.Body(&newUser); err != nil {
		return c.JSON(400, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Validate
	if newUser.Name == "" || newUser.Email == "" {
		return c.JSON(400, map[string]interface{}{
			"error": "Name and email are required",
		})
	}

	// Assign ID (in a real app, this would come from the database)
	newUser.ID = len(users) + 1
	users[newUser.ID] = newUser

	return c.JSON(201, map[string]interface{}{
		"message": "User created successfully",
		"user":    newUser,
	})
}

// handlePanic demonstrates the Recovery middleware
func handlePanic(c *context.Context) error {
	panic("This is a test panic!")
}
