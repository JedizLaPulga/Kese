package kese

import (
	"fmt"
	"net/http"

	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/router"
)

// HandlerFunc defines the function signature for route handlers.
// It receives a Context and returns an error for centralized error handling.
type HandlerFunc func(*context.Context) error

// App is the main application instance that holds the router and configuration.
// It provides a high-level API for defining routes and middleware.
type App struct {
	router     *router.Router
	middleware []MiddlewareFunc
}

// MiddlewareFunc defines the function signature for middleware.
// Middleware can modify the context or terminate the request chain.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// New creates a new Kese application instance.
// This is the starting point for building your web application.
func New() *App {
	return &App{
		router:     router.New(),
		middleware: make([]MiddlewareFunc, 0),
	}
}

// Use adds middleware to the application.
// Middleware is executed in the order it is registered.
func (a *App) Use(middleware ...MiddlewareFunc) {
	a.middleware = append(a.middleware, middleware...)
}

// GET registers a route that responds to GET requests.
func (a *App) GET(path string, handler HandlerFunc) {
	a.addRoute(http.MethodGet, path, handler)
}

// POST registers a route that responds to POST requests.
func (a *App) POST(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPost, path, handler)
}

// PUT registers a route that responds to PUT requests.
func (a *App) PUT(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPut, path, handler)
}

// DELETE registers a route that responds to DELETE requests.
func (a *App) DELETE(path string, handler HandlerFunc) {
	a.addRoute(http.MethodDelete, path, handler)
}

// PATCH registers a route that responds to PATCH requests.
func (a *App) PATCH(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPatch, path, handler)
}

// OPTIONS registers a route that responds to OPTIONS requests.
func (a *App) OPTIONS(path string, handler HandlerFunc) {
	a.addRoute(http.MethodOptions, path, handler)
}

// HEAD registers a route that responds to HEAD requests.
func (a *App) HEAD(path string, handler HandlerFunc) {
	a.addRoute(http.MethodHead, path, handler)
}

// addRoute is the internal method for registering routes with the router.
func (a *App) addRoute(method, path string, handler HandlerFunc) {
	// Wrap the handler with all registered middleware
	wrappedHandler := a.wrapMiddleware(handler)
	a.router.Add(method, path, wrappedHandler)
}

// wrapMiddleware wraps a handler with all registered middleware.
// Middleware is applied in reverse order so that the first registered
// middleware is the outermost layer.
func (a *App) wrapMiddleware(handler HandlerFunc) HandlerFunc {
	// Apply middleware in reverse order
	for i := len(a.middleware) - 1; i >= 0; i-- {
		handler = a.middleware[i](handler)
	}
	return handler
}

// ServeHTTP implements http.Handler interface.
// This allows the App to be used directly with http.Server.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a new context for this request
	ctx := context.New(w, r)

	// Find the matching route
	handler, params := a.router.Match(r.Method, r.URL.Path)
	if handler == nil {
		// No route matched - return 404
		ctx.Status(http.StatusNotFound)
		ctx.String(http.StatusNotFound, "404 Not Found")
		return
	}

	// Set route parameters in context
	ctx.SetParams(params)

	// Execute the handler
	if err := handler(ctx); err != nil {
		// Handle errors returned by handlers
		// For now, we'll just return a 500 error
		// In the future, we can add custom error handlers
		ctx.Status(http.StatusInternalServerError)
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("Internal Server Error: %v", err))
	}
}

// Run starts the HTTP server on the specified address.
// address should be in the format ":8080" or "localhost:8080"
func (a *App) Run(address string) error {
	fmt.Printf("🚀 Kese server starting on %s\n", address)
	return http.ListenAndServe(address, a)
}

// RunTLS starts the HTTPS server on the specified address with TLS config.
func (a *App) RunTLS(address, certFile, keyFile string) error {
	fmt.Printf("🔒 Kese server starting on %s (TLS)\n", address)
	return http.ListenAndServeTLS(address, certFile, keyFile, a)
}
