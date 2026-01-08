package main

import (
	"log"
	"os"

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
	app.Use(middleware.Logger(app.Logger))
	app.Use(middleware.Recovery(app.Logger))
	app.Use(middleware.CORS())
	app.Use(middleware.RequestID())

	// Serve the beautiful landing page at root
	app.GET("/", serveLandingPage)

	// Serve static files (logo image)
	app.StaticFile("/img/kese.png", "../../img/kese.png")

	// API routes (prefixed with /api for clarity)
	app.GET("/api/health", handleHealth)
	app.GET("/api/users", handleGetUsers)
	app.GET("/api/users/:id", handleGetUser)
	app.POST("/api/users", handleCreateUser)

	// Example of error handling
	app.GET("/api/panic", handlePanic)

	// Start server
	log.Println("Starting Kese example server...")
	log.Println("🌐 Navigate to http://localhost:8080 to see the landing page")
	if err := app.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// serveLandingPage serves the beautiful HTML landing page
func serveLandingPage(c *context.Context) error {
	htmlContent, err := os.ReadFile("../../templates/welcome.html")
	if err != nil {
		// Fallback to simple message if template not found
		return c.JSON(200, map[string]interface{}{
			"message": "Welcome to Kese Framework! 🚀",
			"version": "0.1.0",
		})
	}
	return c.HTML(200, string(htmlContent))
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
